package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"gopkg.in/resty.v1"
)

const (
	apiTokenRawRegex = `.*(?:\"?api_token\"?):\s*\"([^"]+)\".*`
)

var (
	apiTokenRegex          = regexp.MustCompile(apiTokenRawRegex)
	errorEmojiDoesNotExist = fmt.Errorf("emoji does not exist")
	errorEmojiExists       = fmt.Errorf("emoji already exists")
	errorEmojiNameTaken    = fmt.Errorf("emoji name is already taken")
)

// Client provides a simple interface for interacting with the Slack API.
type Client struct {
	apiToken           string
	backoffStrategy    backoff.BackOff
	CustomizeEmojiPath string
	EmojiAddPath       string
	EmojiAdminListPath string
	EmojiRemovePath    string
	Emojis             map[string]Emoji
	restClient         *resty.Client
	TeamName           string
}

// NewSlackClient instantiates a Slack client to a single team for emoji upload.
func NewSlackClient(slackTeamName, slackCookie string) (client *Client, err error) {
	client = &Client{
		backoffStrategy:    backoff.NewExponentialBackOff(),
		CustomizeEmojiPath: "customize/emoji",
		EmojiAddPath:       "api/emoji.add",
		EmojiAdminListPath: "api/emoji.adminList",
		EmojiRemovePath:    "api/emoji.remove",
		restClient: resty.NewWithClient(
			&http.Client{
				Timeout: 30 * time.Second,
			},
		).SetHeader("Cookie", slackCookie),
		TeamName: slackTeamName,
	}

	client.apiToken, err = client.APIToken()
	if err != nil {
		return nil, errors.Wrapf(err, "retrieving API token failed, client: '%+v'", client)
	}

	client.Emojis, err = client.GetEmojis()
	if err != nil {
		return nil, errors.Wrapf(err, "retrieving emojis failed, client: '%+v'", client)
	}

	return client, nil
}

// APIToken retrieves and returns the API token from an emoji customization
// call.
func (client *Client) APIToken() (apiToken string, err error) {
	if client == nil {
		return "", fmt.Errorf("client is nil")
	}

	request := client.restClient.R()
	innerError := (error)(nil)
	response := (*resty.Response)(nil)
	err = backoff.RetryNotifyWithTimer(
		func() (err error) {
			response, err = request.Get(client.CustomizeEmojiURI())
			if err != nil {
				requestDump, _ := httputil.DumpRequest(request.RawRequest, true)
				innerError = errors.Wrapf(err, "request failed, request dump: '%+v'", string(requestDump))

				return innerError
			}

			if response.StatusCode() >= 400 &&
				response.StatusCode() < 500 {
				responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
				_ = response.RawResponse.Body.Close()
				innerError = errors.Wrapf(err, "response contains client error, response dump: '%+v'", string(responseDump))

				return innerError
			} else if response.StatusCode() >= 500 &&
				response.StatusCode() < 600 {
				responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
				_ = response.RawResponse.Body.Close()
				innerError = errors.Wrapf(err, "response contains server error, response dump: '%+v'", string(responseDump))

				return innerError
			}

			return nil
		},
		client.backoffStrategy,
		func(err error, backoffDelay time.Duration) {
			log.Printf("requesting API token temporarily failed and will be retried, error: '%+v', backoff delay: '%+v'\n", err, backoffDelay)
		},
		nil,
	)

	defer func() { _ = response.RawResponse.Body.Close() }()

	if err != nil {
		return "", innerError
	}

	bodyReader := bytes.NewReader(response.Body())
	document, err := html.Parse(bodyReader)
	if err != nil {
		innerError = errors.Wrapf(err, "parsing API token body response failed, raw body: '%+v'", string(response.Body()))

		return "", innerError
	}

	apiToken = apiTokenFromHTMLRecursively(document)
	if apiToken == "" {
		innerError = fmt.Errorf("API token not found, raw document: '%+v'", string(response.Body()))

		return "", innerError
	}

	return apiToken, nil
}

// CustomizeEmojiURI returns the URI of the customize/emoji endpoint.
func (client *Client) CustomizeEmojiURI() (uri string) {
	if client == nil {
		return uri
	}

	return client.Host() + "/" + client.CustomizeEmojiPath
}

// DeleteEmoji deletes a single emoji identified by its name from the connected
// Slack team's custom emojis.
func (client *Client) DeleteEmoji(emojiName string) (err error) {
	if client == nil {
		return fmt.Errorf("client is nil")
	}

	if _, isExisting := client.Emojis[emojiName]; !isExisting {
		return errorEmojiDoesNotExist
	}

	innerError := (error)(nil)
	isAssertable := false
	isSuccessful := false
	request := client.restClient.R().
		SetFormData(
			map[string]string{
				"name":  emojiName,
				"token": client.apiToken,
			},
		)
	response := (*resty.Response)(nil)
	responseJSON := make(map[string]interface{})

	err = backoff.RetryNotifyWithTimer(
		func() (err error) {
			response, err = request.Post(client.EmojiRemoveURI())
			if err != nil {
				requestDump, _ := httputil.DumpRequest(request.RawRequest, true)
				innerError = errors.Wrapf(err, "request failed, request dump: '%+v'", string(requestDump))

				return innerError
			}

			defer func() { _ = response.RawResponse.Body.Close() }()

			if response.StatusCode() >= 400 &&
				response.StatusCode() < 500 {
				responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
				innerError = errors.Wrapf(err, "response contains client error, response dump: '%+v'", string(responseDump))

				return innerError
			} else if response.StatusCode() >= 500 &&
				response.StatusCode() < 600 {
				responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
				innerError = errors.Wrapf(err, "response contains server error, response dump: '%+v'", string(responseDump))

				return innerError
			}

			err = json.Unmarshal(response.Body(), &responseJSON)
			if err != nil {
				innerError = errors.Wrapf(err, "unmarshalling JSON response failed, raw JSON response: '%+v'", string(response.Body()))

				return innerError
			}

			isSuccessful, isAssertable = responseJSON["ok"].(bool)
			if !isAssertable {
				innerError = fmt.Errorf("response OK flag could not be asserted to boolean, raw response JSON: '%+v'", string(response.Body()))

				return innerError
			}

			if !isSuccessful {
				innerError = fmt.Errorf("response contained not OK status, response: '%+v'", responseJSON)

				return innerError
			}

			return nil
		},
		client.backoffStrategy,
		func(err error, backoffDelay time.Duration) {
			log.Printf("requesting emoji removal temporarily failed and will be retried, name: '%+v', error: '%+v', backoff delay: '%+v'\n", emojiName, err, backoffDelay)
		},
		nil,
	)
	if err != nil {
		return innerError
	}

	return nil
}

// DeleteEmojis deletes all custom emojis from the connected Slack team.
func (client *Client) DeleteEmojis() (err error) {
	if client == nil {
		return fmt.Errorf("client is nil")
	}

	totalCount := len(client.Emojis)
	deleteCount := 0
	for name := range client.Emojis {
		log.Printf("%s\n", name)

		err = client.DeleteEmoji(name)
		if err != nil &&
			err != errorEmojiDoesNotExist {
			return errors.Wrapf(err, "deleting emoji failed, name: '%+v'", name)
		} else if err == nil {
			log.Printf("deleted\n")
			deleteCount++
		}

		log.Printf("Deleted: %d (%.2f%%), Remaining: %d (%.2f%%), total: %d\n\n", deleteCount, float64(deleteCount)/float64(totalCount)*100.0, totalCount-deleteCount, float64(totalCount-deleteCount)/float64(totalCount)*100.0, totalCount)
	}

	return nil
}

// EmojiAddURI returns the URI of the api/emoji.add endpoint.
func (client *Client) EmojiAddURI() (uri string) {
	if client == nil {
		return uri
	}

	return client.Host() + "/" + client.EmojiAddPath
}

// EmojiAdminListURI returns the URI of the api/emoji.adminList endpoint.
func (client *Client) EmojiAdminListURI() (uri string) {
	if client == nil {
		return uri
	}

	return client.Host() + "/" + client.EmojiAdminListPath
}

// EmojiRemoveURI returns the URI of the api/emoji.remove endpoint.
func (client *Client) EmojiRemoveURI() (uri string) {
	if client == nil {
		return uri
	}

	return client.Host() + "/" + client.EmojiRemovePath
}

// GetEmojis returns the available custom emojis by name in a Slack team.
func (client *Client) GetEmojis() (emojis map[string]Emoji, err error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	emojis = make(map[string]Emoji)
	innerError := (error)(nil)
	page := 1
	pageCount := 2
	pageSize := 1000
	request := client.restClient.R().
		SetFormData(
			map[string]string{
				"count": fmt.Sprintf("%d", pageSize),
				"page":  fmt.Sprintf("%d", page),
				"query": "",
				"token": client.apiToken,
			},
		)
	response := (*resty.Response)(nil)
	responseJSON := EmojiListResponse{}

	for page <= pageCount {
		request.FormData["page"][0] = fmt.Sprintf("%d", page)
		request.FormData["count"][0] = fmt.Sprintf("%d", pageSize)

		err = backoff.RetryNotifyWithTimer(
			func() (err error) {
				response, err = request.Post(client.EmojiAdminListURI())
				if err != nil {
					requestDump, _ := httputil.DumpRequest(request.RawRequest, true)

					return errors.Wrapf(err, "request failed, request dump: '%+v'", string(requestDump))
				}

				defer func() { _ = response.RawResponse.Body.Close() }()

				if response.StatusCode() >= 400 &&
					response.StatusCode() < 500 {
					responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
					innerError = errors.Wrapf(err, "response contains client error, response dump: '%+v'", string(responseDump))

					return innerError
				} else if response.StatusCode() >= 500 &&
					response.StatusCode() < 600 {
					responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
					innerError = errors.Wrapf(err, "response contains server error, response dump: '%+v'", string(responseDump))

					return innerError
				}

				err = json.Unmarshal(response.Body(), &responseJSON)
				if err != nil {
					innerError = errors.Wrapf(err, "unmarshalling JSON response failed, raw JSON response: '%+v'", string(response.Body()))

					return innerError
				}

				if !responseJSON.IsOk {
					innerError = fmt.Errorf("not OK response received, response: '%+v'", responseJSON)

					return innerError
				}

				return nil
			},
			client.backoffStrategy,
			func(err error, backoffDelay time.Duration) {
				log.Printf("requesting emoji list temporarily failed and will be retried, error: '%+v', backoff delay: '%+v'\n", err, backoffDelay)
			},
			nil,
		)
		if err != nil {
			return nil, innerError
		}

		for _, emoji := range responseJSON.Emojis {
			emojis[emoji.Name] = emoji
		}

		pageCount = responseJSON.Paging.PageCount
		page = responseJSON.Paging.Page + 1
	}

	return emojis, nil
}

// Host returns the Slack host URL for the configured team.
func (client *Client) Host() (host string) {
	if client == nil {
		return host
	}

	return fmt.Sprintf("https://%s.slack.com", client.TeamName)
}

// PostEmoji uploads an emoji file specified with its path under the given name.
func (client *Client) PostEmoji(emojiName, emojiPath string) (err error) {
	if client == nil {
		return fmt.Errorf("client is nil")
	}

	if _, isExisting := client.Emojis[emojiName]; isExisting {
		return errorEmojiExists
	}

	innerError := (error)(nil)
	isAssertable := false
	isSuccessful := false
	request := client.restClient.R().
		SetFormData(
			map[string]string{
				"mode":  "data",
				"name":  emojiName,
				"token": client.apiToken,
			},
		).
		SetFile("image", emojiPath)
	response := (*resty.Response)(nil)
	responseJSON := make(map[string]interface{})

	err = backoff.RetryNotifyWithTimer(
		func() (err error) {
			response, err = request.Post(client.EmojiAddURI())
			if err != nil {
				requestDump, _ := httputil.DumpRequest(request.RawRequest, true)
				innerError = errors.Wrapf(err, "request failed, request dump: '%+v'", string(requestDump))

				return innerError
			}

			defer func() { _ = response.RawResponse.Body.Close() }()

			if response.StatusCode() == 429 {
				_ = response.RawResponse.Body.Close()
				retrySeconds := response.Header().Get("Retry-After")
				retryDuration, err := time.ParseDuration(retrySeconds + "s")
				if err != nil {
					innerError = errors.Wrapf(err, "parsing retry duration failed, raw retry seconds: '%+v'", retrySeconds)

					return innerError
				}

				log.Printf("waiting rate limit for %s\n", retryDuration)
				<-time.After(retryDuration)

				response, err = request.Post(client.EmojiAddURI())
				if err != nil {
					requestDump, _ := httputil.DumpRequest(request.RawRequest, true)
					innerError = errors.Wrapf(err, "rate limited request failed, request dump: '%+v'", string(requestDump))

					return innerError
				}

				if response.StatusCode() >= 400 &&
					response.StatusCode() < 500 {
					responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
					innerError = errors.Wrapf(err, "rate limited response contains client error, response dump: '%+v'", string(responseDump))

					return innerError
				} else if response.StatusCode() >= 500 &&
					response.StatusCode() < 600 {
					responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
					innerError = errors.Wrapf(err, "rate limited response contains server error, response dump: '%+v'", string(responseDump))

					return innerError
				}
			} else if response.StatusCode() >= 400 &&
				response.StatusCode() < 500 {
				responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
				innerError = errors.Wrapf(err, "response contains client error, response dump: '%+v'", string(responseDump))

				return innerError
			} else if response.StatusCode() >= 500 &&
				response.StatusCode() < 600 {
				responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
				innerError = errors.Wrapf(err, "response contains server error, response dump: '%+v'", string(responseDump))

				return innerError
			}

			err = json.Unmarshal(response.Body(), &responseJSON)
			if err != nil {
				innerError = errors.Wrapf(err, "unmarshalling JSON response failed, raw JSON response: '%+v'", string(response.Body()))

				return innerError
			}

			isSuccessful, isAssertable = responseJSON["ok"].(bool)
			if !isAssertable {
				innerError = fmt.Errorf("response OK flag could not be asserted to boolean, raw response JSON: '%+v'", string(response.Body()))

				return innerError
			}

			if !isSuccessful {
				errorString, isAssertable := responseJSON["error"].(string)
				if isAssertable &&
					(errorString == "error_name_taken" ||
						errorString == "error_name_taken_i18n") {
					innerError = errorEmojiNameTaken

					return backoff.Permanent(innerError)
				}

				innerError = fmt.Errorf("retrying request, failed response: '%+v'", responseJSON)

				return innerError
			}

			return nil
		},
		client.backoffStrategy,
		func(err error, backoffDelay time.Duration) {
			log.Printf("requesting emoji addition temporarily failed and will be retried, error: '%+v', backoff delay: '%+v'\n", err, backoffDelay)
		},
		nil,
	)
	if err != nil {
		return innerError
	}

	return nil
}

// PostEmojis uploads all emojis in the specified directory using the file's
// name without extension as the emoji name prefixed and suffixed with the
// specified qualifiers.
func (client *Client) PostEmojis(emojiDirectoryPath, emojiAliasPrefix, emojiAliasSuffix, emojiAliasTakenPrefix, emojiAliasTakenSuffix string) (err error) {
	if client == nil {
		return fmt.Errorf("client is nil")
	} else if emojiDirectoryPath == "" {
		return fmt.Errorf("invalid empty emoji directory path")
	} else if emojiAliasTakenSuffix == "" {
		return fmt.Errorf("invalid empty emoji alias taken suffix")
	}

	skipCount := 0
	totalCount := 0
	uploadCount := 0

	err = filepath.Walk(emojiDirectoryPath, func(path string, info os.FileInfo, itemError error) (walkError error) {
		if itemError != nil {
			return errors.Wrapf(err, "walking path failed, path: '%+v', info: '%+v'", path, info)
		} else if info.IsDir() {
			return nil
		}

		totalCount++

		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "iterating emoji directory for counting failed, emoji directory path: '%+v'", emojiDirectoryPath)
	}

	err = filepath.Walk(emojiDirectoryPath, func(path string, info os.FileInfo, itemError error) (walkError error) {
		if itemError != nil {
			return errors.Wrapf(err, "walking path failed, path: '%+v', info: '%+v'", path, info)
		} else if info.IsDir() {
			return nil
		}

		name, takenName := newEmojiNameFromFilePath(path, emojiAliasPrefix, emojiAliasSuffix, emojiAliasTakenPrefix, emojiAliasTakenSuffix)
		log.Printf("%s\n", filepath.Base(path))
		log.Printf("sanitized prefixed and suffixed name: %+v\n", name)

		err = client.PostEmoji(name, path)
		if err != nil &&
			err != errorEmojiExists &&
			err != errorEmojiNameTaken {
			return errors.Wrapf(err, "posting emoji failed, path: '%+v'", path)
		} else if err != nil &&
			err == errorEmojiExists {
			log.Printf("skipped existing\n")
			skipCount++
		} else if err != nil &&
			err == errorEmojiNameTaken {
			log.Printf("name is taken by non-custom emoji, using taken prefixed and suffixed name: %+v", takenName)
			err = client.PostEmoji(takenName, path)
			if err != nil &&
				err != errorEmojiExists &&
				err != errorEmojiNameTaken {
				return errors.Wrapf(err, "posting emoji failed, path: '%+v''", path)
			} else if err != nil &&
				err == errorEmojiExists {
				log.Printf("skipped existing\n")
				skipCount++
			} else if err != nil &&
				err == errorEmojiNameTaken {
				return fmt.Errorf("original and taken names were already taken, taken name: '%+v'", takenName)
			} else if err == nil {
				log.Printf("uploaded")
				uploadCount++
			}
		} else if err == nil {
			log.Printf("uploaded")
			uploadCount++
		}

		log.Printf("Skipped+Uploaded=Existing: %d+%d=%d (%.2f%%+%.2f%%=%.2f%%), Remaining: %d (%.2f%%), total: %d\n\n", skipCount, uploadCount, skipCount+uploadCount, float64(skipCount)/float64(totalCount)*100.0, float64(uploadCount)/float64(totalCount)*100.0, float64(skipCount+uploadCount)/float64(totalCount)*100.0, totalCount-(skipCount+uploadCount), float64(totalCount-(skipCount+uploadCount))/float64(totalCount)*100.0, totalCount)

		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "iterating emoji directory for uploading failed, emoji directory path: '%+v'", emojiDirectoryPath)
	}

	return nil
}

// apiTokenFromHTMLRecursively takes a customize/emoji HTML response and parses
// the API token out of it.
func apiTokenFromHTMLRecursively(node *html.Node) (apiToken string) {
	if node.Parent != nil &&
		node.Parent.Parent != nil &&
		node.Parent.Parent.Parent != nil &&
		node.Parent.Parent.Parent.Parent != nil &&
		node.Parent.Parent.Parent.Parent.Type == html.DocumentNode &&
		node.Parent.Parent.Parent.Parent.Data == "" &&
		node.Parent.Parent.Parent.Type == html.ElementNode &&
		node.Parent.Parent.Parent.Data == "html" &&
		node.Parent.Parent.Type == html.ElementNode &&
		node.Parent.Parent.Data == "body" &&
		node.Parent.Type == html.ElementNode &&
		node.Parent.Data == "script" &&
		len(node.Parent.Attr) != 0 &&
		node.Parent.Attr[0].Key == "type" &&
		node.Parent.Attr[0].Val == "text/javascript" &&
		node.Type == html.TextNode &&
		strings.Contains(node.Data, "api_token") {
		groups := apiTokenRegex.FindStringSubmatch(node.Data)
		if len(groups) == 2 {
			return groups[1]
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		apiToken = apiTokenFromHTMLRecursively(child)

		if apiToken != "" {
			return apiToken
		}
	}

	return ""
}

// newEmojiNameFromFilePath returns the prefixed and suffixed name and taken name from the
// emoji's file path.
func newEmojiNameFromFilePath(path, prefix, suffix, takenPrefix, takenSuffix string) (name, takenName string) {
	fileName := filepath.Base(path)
	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	sanitizedName := strings.Replace(baseName, ":", "", -1) // Note: for some reason on macOS path of /73.jpg is read as :73.jpg. : is not permitted anyway, because it denotes emoji open/close tags.
	name = prefix + sanitizedName + suffix

	return name, takenPrefix + name + takenSuffix
}
