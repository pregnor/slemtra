package slack

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http/httputil"
	"path"
	"regexp"
	"strings"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/pregnor/slemtra/pkg/errors"
	"github.com/pregnor/slemtra/pkg/slack/model"
	"golang.org/x/net/html"
	"gopkg.in/resty.v1"
)

// EmojiService provides methods for interacting with the Slack API.
type EmojiService interface {
	// CreateEmojiAlias adds aliases for existing emojis to the workspace.
	// CreateEmojiAliases(aliasModels ...model.CreateEmojisAliasModel) (err error)

	// CreateEmojis adds multiple named emojis to the workspace.
	// CreateEmojis(emojiModels ...model.CreateEmojisModel) (err error)

	// DeleteEmojis removes the named emojis from the workspace. When no names
	// are passed, all the emojis are deleted.
	// DeleteEmojis(emojiModels ...model.DeleteEmojisModel) (err error)

	// DownloadEmojis downloads the graphical content of an emoji from the
	// workspace.
	// DownloadEmojis(emojiModels ...model.DownloadEmojisModel) (err error)

	// ListEmojis returns the available emojis in the workspace. The output is
	// sorted by the emoji names.
	ListEmojis() (emojiModels []model.ListEmojisModel, err error)

	// RenameEmojis renames an existing emoji in the workspace.
	// RenameEmojis(emojiModels ...model.RenameEmojisModel) (err error)
}

// NewEmojiService instantiates a Slack emoji service to a single team for emoji
// operations.
// func NewEmojiService(workspaceName, slackCookie string) (service EmojiService, err error) {
// 	clientImplementation := &client{
// 		backoffStrategy: backoff.NewExponentialBackOff(),
// 		restClient: resty.NewWithClient(
// 			&http.Client{
// 				Timeout: 30 * time.Second,
// 			},
// 		).SetHeader("Cookie", slackCookie),
// 		workspaceName: fmt.Sprintf("https://%s.slack.com", workspaceName),
// 	}

// 	clientImplementation.apiToken, err = service.APIToken()
// 	if err != nil {
// 		return nil, errors.Wrapf(err, "retrieving API token failed, client: '%+v'", clientImplementation)
// 	}

// 	clientImplementation.Emojis, err = service.GetEmojis()
// 	if err != nil {
// 		return nil, errors.Wrapf(err, "retrieving emojis failed, client: '%+v'", clientImplementation)
// 	}

// 	return clientImplementation, nil
// }.

// emojiService provides an implementation for the emoji service.
//
// Implements the slack.EmojiService interface.
type emojiService struct {
	apiToken        string
	backoffStrategy backoff.BackOff
	restClient      *resty.Client
	workspaceHost   string
}

// APIToken retrieves and returns the API token from an emoji customization
// call.
// func (service *emojiService) APIToken() (apiToken string, err error) {
// 	if client == nil {
// 		return "", fmt.Errorf("client is nil")
// 	}

// 	request := service.restClient.R()
// 	innerError := (error)(nil)
// 	response := (*resty.Response)(nil)
// 	err = backoff.RetryNotifyWithTimer(
// 		func() (err error) {
// 			response, err = request.Get(service.CustomizeEmojiURI())
// 			if err != nil {
// 				requestDump, _ := httputil.DumpRequest(request.RawRequest, true)
// 				innerError = errors.Wrapf(err, "request failed, request dump: '%+v'", string(requestDump))

// 				return innerError
// 			}

// 			if response.StatusCode() >= 400 &&
// 				response.StatusCode() < 500 {
// 				responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
// 				_ = response.RawResponse.Body.Close()
// 				innerError = errors.Wrapf(err, "response contains client error, response dump: '%+v'", string(responseDump))

// 				return innerError
// 			} else if response.StatusCode() >= 500 &&
// 				response.StatusCode() < 600 {
// 				responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
// 				_ = response.RawResponse.Body.Close()
// 				innerError = errors.Wrapf(err, "response contains server error, response dump: '%+v'", string(responseDump))

// 				return innerError
// 			}

// 			return nil
// 		},
// 		service.backoffStrategy,
// 		func(err error, backoffDelay time.Duration) {
// 			log.Printf("requesting API token temporarily failed and will be retried, error: '%+v', backoff delay: '%+v'\n", err, backoffDelay)
// 		},
// 		nil,
// 	)

// 	defer func() { _ = response.RawResponse.Body.Close() }()

// 	if err != nil {
// 		return "", innerError
// 	}

// 	bodyReader := bytes.NewReader(response.Body())
// 	document, err := html.Parse(bodyReader)
// 	if err != nil {
// 		innerError = errors.Wrapf(err, "parsing API token body response failed, raw body: '%+v'", string(response.Body()))

// 		return "", innerError
// 	}

// 	apiToken = apiTokenFromHTMLRecursively(document)
// 	if apiToken == "" {
// 		innerError = fmt.Errorf("API token not found, raw document: '%+v'", string(response.Body()))

// 		return "", innerError
// 	}

// 	return apiToken, nil
// }.

// DeleteEmoji deletes a single emoji identified by its name from the connected
// Slack team's custom emojis.
// func (service *emojiService) DeleteEmoji(emojiName string) (err error) {
// 	if client == nil {
// 		return fmt.Errorf("client is nil")
// 	}

// 	if _, isExisting := service.Emojis[emojiName]; !isExisting {
// 		return errorEmojiDoesNotExist
// 	}

// 	innerError := (error)(nil)
// 	isAssertable := false
// 	isSuccessful := false
// 	request := service.restClient.R().
// 		SetFormData(
// 			map[string]string{
// 				"name":  emojiName,
// 				"token": service.apiToken,
// 			},
// 		)
// 	response := (*resty.Response)(nil)
// 	responseJSON := make(map[string]interface{})

// 	err = backoff.RetryNotifyWithTimer(
// 		func() (err error) {
// 			response, err = request.Post(service.EmojiRemoveURI())
// 			if err != nil {
// 				requestDump, _ := httputil.DumpRequest(request.RawRequest, true)
// 				innerError = errors.Wrapf(err, "request failed, request dump: '%+v'", string(requestDump))

// 				return innerError
// 			}

// 			defer func() { _ = response.RawResponse.Body.Close() }()

// 			if response.StatusCode() >= 400 &&
// 				response.StatusCode() < 500 {
// 				responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
// 				innerError = errors.Wrapf(err, "response contains client error, response dump: '%+v'", string(responseDump))

// 				return innerError
// 			} else if response.StatusCode() >= 500 &&
// 				response.StatusCode() < 600 {
// 				responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
// 				innerError = errors.Wrapf(err, "response contains server error, response dump: '%+v'", string(responseDump))

// 				return innerError
// 			}

// 			err = json.Unmarshal(response.Body(), &responseJSON)
// 			if err != nil {
// 				innerError = errors.Wrapf(err, "unmarshalling JSON response failed, raw JSON response: '%+v'", string(response.Body()))

// 				return innerError
// 			}

// 			isSuccessful, isAssertable = responseJSON["ok"].(bool)
// 			if !isAssertable {
// 				innerError = fmt.Errorf("response OK flag could not be asserted to boolean, raw response JSON: '%+v'", string(response.Body()))

// 				return innerError
// 			}

// 			if !isSuccessful {
// 				innerError = fmt.Errorf("response contained not OK status, response: '%+v'", responseJSON)

// 				return innerError
// 			}

// 			return nil
// 		},
// 		service.backoffStrategy,
// 		func(err error, backoffDelay time.Duration) {
// 			log.Printf("requesting emoji removal temporarily failed and will be retried, name: '%+v', error: '%+v', backoff delay: '%+v'\n", emojiName, err, backoffDelay)
// 		},
// 		nil,
// 	)
// 	if err != nil {
// 		return innerError
// 	}

// 	return nil
// }.

// DeleteEmojis removes the named emojis from the workspace. When no names
// are passed, all the emojis are deleted.
//
// Implements the slack.EmojiService interface.
// func (service *emojiService) DeleteEmojis(emojiModels ...model.DeleteEmojisModel) (err error) {
// 	if service == nil {
// 		return errors.NewErrorWithDetails(errors.ErrorInvalidValue,
// 			"context", "emojiService.DeleteEmojis", "key", "service", "reason", "service is nil")
// 	} else if len(emojiModels) == 0 {
// 		emojiList, err := service.ListEmojis()
// 		if err != nil {
// 			return errors.NewErrorWithDetails(errors.ErrorOperationFailed,
// 				"context", "emojiService.DeleteEmojis", "error", err, "operation", "service.ListEmojis()")
// 		}

// 		emojiDeletionList := make([]model.DeleteEmojisModel, 0, len(emojiList))
// 		for _, emoji := range emojiList {
// 			emojiDeletionList = append(emojiDeletionList, model.NewDeleteEmojisModel(emoji.Name()))
// 		}

// 		return service.DeleteEmojis(emojiDeletionList...)
// 	}

// 	totalCount := len(service.Emojis)
// 	deleteCount := 0
// 	for name := range service.Emojis {
// 		log.Printf("%s\n", name)

// 		err = service.DeleteEmoji(name)
// 		if err != nil &&
// 			err != errorEmojiDoesNotExist {
// 			return errors.Wrapf(err, "deleting emoji failed, name: '%+v'", name)
// 		} else if err == nil {
// 			log.Printf("deleted\n")
// 			deleteCount++
// 		}

// 		log.Printf("Deleted: %d (%.2f%%), Remaining: %d (%.2f%%), total: %d\n\n", deleteCount, float64(deleteCount)/float64(totalCount)*100.0, totalCount-deleteCount, float64(totalCount-deleteCount)/float64(totalCount)*100.0, totalCount)
// 	}

// 	return nil
// }.

// ListEmojis returns the available emojis in the workspace. The output is
// sorted by the emoji names.
//
// Implements the slack.EmojiService interface.
func (service *emojiService) ListEmojis() (emojiModels []model.ListEmojisModel, err error) {
	if service == nil {
		return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
			"context", "emojiService.ListEmojis", "key", "service", "reason", "service is nil")
	}

	page := 1
	pageCount := 1 // Note: initial dummy value to start the paging.
	pageSize := 1000
	request := service.restClient.R().
		SetFormData(
			map[string]string{
				"count": fmt.Sprintf("%d", pageSize),
				"page":  fmt.Sprintf("%d", page),
				"query": "",
				"token": service.apiToken,
			},
		)
	requestURL := service.URL("api", "admin.emoji.list")
	response := (*resty.Response)(nil)
	responseModel := model.ListEmojisResponse{}

	for page <= pageCount {
		request.FormData["page"][0] = fmt.Sprintf("%d", page)
		request.FormData["count"][0] = fmt.Sprintf("%d", pageSize)

		err = backoff.RetryNotifyWithTimer(
			func() (err error) {
				response, err = request.Post(requestURL)
				if err != nil {
					requestDump, _ := httputil.DumpRequest(request.RawRequest, true)
					var responseDump []byte
					if response != nil &&
						response.RawResponse != nil {
						responseDump, _ = httputil.DumpResponse(response.RawResponse, true)
					}

					return errors.NewErrorWithDetails(errors.ErrorOperationFailed,
						"context", "emojiService.ListEmojis", "error", err, "operation", "POST "+requestURL,
						"reason", "request failed", "request", string(requestDump), "response", string(responseDump))
				}

				defer func() { _ = response.RawResponse.Body.Close() }()

				if response.StatusCode() >= 400 &&
					response.StatusCode() < 600 {
					reason := "response server error"
					if response.StatusCode() < 500 { // && response.StatusCode() >= 400 {
						reason = "response client error"
					}
					responseDump, _ := httputil.DumpResponse(response.RawResponse, true)

					return errors.NewErrorWithDetails(errors.ErrorOperationFailed,
						"context", "emojiService.ListEmojis", "error", err, "operation", "POST "+requestURL,
						"reason", reason, "response", string(responseDump))
				}

				err = json.Unmarshal(response.Body(), &responseModel)
				if err != nil {
					return errors.NewErrorWithDetails(errors.ErrorInvalidValue,
						"context", "emojiService.ListEmojis", "error", err, "key", "response.Body()",
						"reason", "parsing response failed", "value", string(response.Body()))
				}

				if !responseModel.IsOk {
					return errors.NewErrorWithDetails(errors.ErrorInvalidValue,
						"context", "emojiService.ListEmojis", "key", "responseModel",
						"reason", "not OK response received", "value", responseModel)
				}

				return nil
			},
			service.backoffStrategy,
			notifyBackOffRetry,
			nil,
		)
		if err != nil {
			return nil, errors.NewErrorWithDetails(errors.ErrorOperationFailed,
				"context", "emojiService.ListEmojis", "error", err, "operation", "retryWithBackoff",
				"reason", "backoff retry operation failed")
		}

		if emojiModels == nil {
			emojiModels = make([]model.ListEmojisModel, 0, responseModel.Paging.TotalItemCount)
		}

		var emojiModel model.ListEmojisModel
		for emojiIndex, emoji := range responseModel.Emojis {
			emojiModel, err = model.NewListEmojisModel(&responseModel.Emojis[emojiIndex])
			if err != nil {
				return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
					"context", "emojiService.ListEmojis", "key", "emojiResponse", "reason", "creating emoji model failed", "value", emoji)
			}

			emojiModels = append(emojiModels, emojiModel)
		}

		pageCount = responseModel.Paging.PageCount
		page = responseModel.Paging.Page + 1
	}

	return emojiModels, nil
}

// PostEmoji uploads an emoji file specified with its path under the given name.
// func (service *emojiService) PostEmoji(emojiName, emojiPath string) (err error) {
// 	if client == nil {
// 		return fmt.Errorf("client is nil")
// 	}

// 	if _, isExisting := service.Emojis[emojiName]; isExisting {
// 		return errorEmojiExists
// 	}

// 	innerError := (error)(nil)
// 	isAssertable := false
// 	isSuccessful := false
// 	request := service.restClient.R().
// 		SetFormData(
// 			map[string]string{
// 				"mode":  "data",
// 				"name":  emojiName,
// 				"token": service.apiToken,
// 			},
// 		).
// 		SetFile("image", emojiPath)
// 	response := (*resty.Response)(nil)
// 	responseJSON := make(map[string]interface{})

// 	err = backoff.RetryNotifyWithTimer(
// 		func() (err error) {
// 			response, err = request.Post(service.EmojiAddURI())
// 			if err != nil {
// 				requestDump, _ := httputil.DumpRequest(request.RawRequest, true)
// 				innerError = errors.Wrapf(err, "request failed, request dump: '%+v'", string(requestDump))

// 				return innerError
// 			}

// 			defer func() { _ = response.RawResponse.Body.Close() }()

// 			if response.StatusCode() == 429 {
// 				_ = response.RawResponse.Body.Close()
// 				retrySeconds := response.Header().Get("Retry-After")
// 				retryDuration, err := time.ParseDuration(retrySeconds + "s")
// 				if err != nil {
// 					innerError = errors.Wrapf(err, "parsing retry duration failed, raw retry seconds: '%+v'", retrySeconds)

// 					return innerError
// 				}

// 				log.Printf("waiting rate limit for %s\n", retryDuration)
// 				<-time.After(retryDuration)

// 				response, err = request.Post(service.EmojiAddURI())
// 				if err != nil {
// 					requestDump, _ := httputil.DumpRequest(request.RawRequest, true)
// 					innerError = errors.Wrapf(err, "rate limited request failed, request dump: '%+v'", string(requestDump))

// 					return innerError
// 				}

// 				if response.StatusCode() >= 400 &&
// 					response.StatusCode() < 500 {
// 					responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
// 					innerError = errors.Wrapf(err, "rate limited response contains client error, response dump: '%+v'", string(responseDump))

// 					return innerError
// 				} else if response.StatusCode() >= 500 &&
// 					response.StatusCode() < 600 {
// 					responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
// 					innerError = errors.Wrapf(err, "rate limited response contains server error, response dump: '%+v'", string(responseDump))

// 					return innerError
// 				}
// 			} else if response.StatusCode() >= 400 &&
// 				response.StatusCode() < 500 {
// 				responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
// 				innerError = errors.Wrapf(err, "response contains client error, response dump: '%+v'", string(responseDump))

// 				return innerError
// 			} else if response.StatusCode() >= 500 &&
// 				response.StatusCode() < 600 {
// 				responseDump, _ := httputil.DumpResponse(response.RawResponse, true)
// 				innerError = errors.Wrapf(err, "response contains server error, response dump: '%+v'", string(responseDump))

// 				return innerError
// 			}

// 			err = json.Unmarshal(response.Body(), &responseJSON)
// 			if err != nil {
// 				innerError = errors.Wrapf(err, "unmarshalling JSON response failed, raw JSON response: '%+v'", string(response.Body()))

// 				return innerError
// 			}

// 			isSuccessful, isAssertable = responseJSON["ok"].(bool)
// 			if !isAssertable {
// 				innerError = fmt.Errorf("response OK flag could not be asserted to boolean, raw response JSON: '%+v'", string(response.Body()))

// 				return innerError
// 			}

// 			if !isSuccessful {
// 				errorString, isAssertable := responseJSON["error"].(string)
// 				if isAssertable &&
// 					(errorString == "error_name_taken" ||
// 						errorString == "error_name_taken_i18n") {
// 					innerError = errorEmojiNameTaken

// 					return backoff.Permanent(innerError)
// 				}

// 				innerError = fmt.Errorf("retrying request, failed response: '%+v'", responseJSON)

// 				return innerError
// 			}

// 			return nil
// 		},
// 		service.backoffStrategy,
// 		func(err error, backoffDelay time.Duration) {
// 			log.Printf("requesting emoji addition temporarily failed and will be retried, error: '%+v', backoff delay: '%+v'\n", err, backoffDelay)
// 		},
// 		nil,
// 	)
// 	if err != nil {
// 		return innerError
// 	}

// 	return nil
// }.

// PostEmojis uploads all emojis in the specified directory using the file's
// name without extension as the emoji name prefixed and suffixed with the
// specified qualifiers.
// func (service *emojiService) PostEmojis(emojiDirectoryPath, emojiAliasPrefix, emojiAliasSuffix, emojiAliasTakenPrefix, emojiAliasTakenSuffix string) (err error) {
// 	if client == nil {
// 		return fmt.Errorf("client is nil")
// 	} else if emojiDirectoryPath == "" {
// 		return fmt.Errorf("invalid empty emoji directory path")
// 	} else if emojiAliasTakenSuffix == "" {
// 		return fmt.Errorf("invalid empty emoji alias taken suffix")
// 	}

// 	skipCount := 0
// 	totalCount := 0
// 	uploadCount := 0

// 	err = filepath.Walk(emojiDirectoryPath, func(path string, info os.FileInfo, itemError error) (walkError error) {
// 		if itemError != nil {
// 			return errors.Wrapf(err, "walking path failed, path: '%+v', info: '%+v'", path, info)
// 		} else if info.IsDir() {
// 			return nil
// 		}

// 		totalCount++

// 		return nil
// 	})
// 	if err != nil {
// 		return errors.Wrapf(err, "iterating emoji directory for counting failed, emoji directory path: '%+v'", emojiDirectoryPath)
// 	}

// 	err = filepath.Walk(emojiDirectoryPath, func(path string, info os.FileInfo, itemError error) (walkError error) {
// 		if itemError != nil {
// 			return errors.Wrapf(err, "walking path failed, path: '%+v', info: '%+v'", path, info)
// 		} else if info.IsDir() {
// 			return nil
// 		}

// 		name, takenName := newEmojiNameFromFilePath(path, emojiAliasPrefix, emojiAliasSuffix, emojiAliasTakenPrefix, emojiAliasTakenSuffix)
// 		log.Printf("%s\n", filepath.Base(path))
// 		log.Printf("sanitized prefixed and suffixed name: %+v\n", name)

// 		err = service.PostEmoji(name, path)
// 		if err != nil &&
// 			err != errorEmojiExists &&
// 			err != errorEmojiNameTaken {
// 			return errors.Wrapf(err, "posting emoji failed, path: '%+v'", path)
// 		} else if err != nil &&
// 			err == errorEmojiExists {
// 			log.Printf("skipped existing\n")
// 			skipCount++
// 		} else if err != nil &&
// 			err == errorEmojiNameTaken {
// 			log.Printf("name is taken by non-custom emoji, using taken prefixed and suffixed name: %+v", takenName)
// 			err = service.PostEmoji(takenName, path)
// 			if err != nil &&
// 				err != errorEmojiExists &&
// 				err != errorEmojiNameTaken {
// 				return errors.Wrapf(err, "posting emoji failed, path: '%+v''", path)
// 			} else if err != nil &&
// 				err == errorEmojiExists {
// 				log.Printf("skipped existing\n")
// 				skipCount++
// 			} else if err != nil &&
// 				err == errorEmojiNameTaken {
// 				return fmt.Errorf("original and taken names were already taken, taken name: '%+v'", takenName)
// 			} else if err == nil {
// 				log.Printf("uploaded")
// 				uploadCount++
// 			}
// 		} else if err == nil {
// 			log.Printf("uploaded")
// 			uploadCount++
// 		}

// 		log.Printf("Skipped+Uploaded=Existing: %d+%d=%d (%.2f%%+%.2f%%=%.2f%%), Remaining: %d (%.2f%%), total: %d\n\n", skipCount, uploadCount, skipCount+uploadCount, float64(skipCount)/float64(totalCount)*100.0, float64(uploadCount)/float64(totalCount)*100.0, float64(skipCount+uploadCount)/float64(totalCount)*100.0, totalCount-(skipCount+uploadCount), float64(totalCount-(skipCount+uploadCount))/float64(totalCount)*100.0, totalCount)

// 		return nil
// 	})
// 	if err != nil {
// 		return errors.Wrapf(err, "iterating emoji directory for uploading failed, emoji directory path: '%+v'", emojiDirectoryPath)
// 	}

// 	return nil
// }.

// URL returns a constructed URL string from the specified path elements and the
// workspace host.
func (service *emojiService) URL(pathElements ...string) (url string) {
	if service == nil {
		return ""
	}

	return fmt.Sprintf("%s/%s", service.workspaceHost, path.Join(pathElements...))
}

// apiTokenFromHTMLRecursively takes a customize/emoji HTML response and parses
// the API token out of it.
func apiTokenFromHTMLRecursively(node *html.Node) (apiToken string) {
	apiTokenRegexp := regexp.MustCompile(`.*(?:\"?api_token\"?):\s*\"([^"]+)\".*`)

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
		groups := apiTokenRegexp.FindStringSubmatch(node.Data)
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

func newBackOffOperation()

// notifyBackOffRetry sens a retry notification to the backoff service when an
// error is encountered which can be retried.
func notifyBackOffRetry(err error, backOffDelay time.Duration) { // nolint:unused // It is passed to the back-off logic.
	log.Printf("backoff retry, error: '%+v', backoff delay: '%+v'\n", err, backOffDelay)
}
