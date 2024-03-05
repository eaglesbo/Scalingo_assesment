package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const githubAPIBaseURL = "https://api.github.com/search/repositories"

type GitHubResponse struct {
	TotalCount int          `json:"total_count"`
	Items      []Repository `json:"items"`
}

type RepositoriesResponse struct {
	TotalCount   int          `json:"total_count"`
	Repositories []Repository `json:"repositories"`
}

type Repository struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	Language    string `json:"language"`
	License     *struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"license"`
	Owner struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"owner"`
	StargazersCount int    `json:"stargazers_count"`
	CreatedAt       string `json:"created_at"`
}

func (handler *Handler) fetchRepositoriesWithFilters(date, language, license, org string, page int) ([]Repository, error) {
	var allRepos []Repository
	var fullURL string

	if org != "" {
		fullURL = fmt.Sprintf("https://api.github.com/orgs/%s/repos?sort=created&type=public&per_page=100&page=%d", org, page)
	} else {
		queryParts := []string{fmt.Sprintf("created:>=%s", date)}
		if language != "" {
			queryParts = append(queryParts, fmt.Sprintf("language:%s", language))
		}
		if license != "" {
			queryParts = append(queryParts, fmt.Sprintf("license:%s", license))
		}
		fullQuery := url.QueryEscape(strings.Join(queryParts, " "))
		fmt.Println(fullQuery)
		fullURL = fmt.Sprintf("%s?q=%s&sort=created&per_page=100&page=%d", githubAPIBaseURL, fullQuery, page)
		fmt.Println(fullURL)
	}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", handler.Config.GetString("Personal_access_token"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned non-200 status: %d", resp.StatusCode)
	}

	if org != "" {
		var orgRepos []Repository
		if err := json.NewDecoder(resp.Body).Decode(&orgRepos); err != nil {
			return nil, err
		}
		allRepos = append(allRepos, orgRepos...)
	} else {

		var gitHubResponse GitHubResponse
		if err := json.NewDecoder(resp.Body).Decode(&gitHubResponse); err != nil {
			return nil, err
		}
		allRepos = append(allRepos, gitHubResponse.Items...)
	}

	return allRepos, nil
}

func (handler *Handler) RepositoriesHandler(w http.ResponseWriter, r *http.Request) {
	if !handler.isValidToken(r) {
		http.Error(w, "Invalid jwt token", http.StatusUnauthorized)
		return
	}
	queryValues := r.URL.Query()
	language := queryValues.Get("language")
	license := queryValues.Get("license")
	org := queryValues.Get("org")

	defaultLimit := 100
	page, err := strconv.Atoi(queryValues.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(queryValues.Get("limit"))
	if err != nil || limit < 1 {
		limit = defaultLimit
	}

	if org != "" {
		limit = defaultLimit
		page = 1
	}

	allRepos, err := handler.fetchAndAccumulateRepos(language, license, org)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, fmt.Sprintf("Failed to fetch repositories: %v", err), http.StatusInternalServerError)
		return
	}

	var startIndex, endIndex int
	startIndex = 0
	endIndex = len(allRepos)
	startIndex = (page - 1) * limit
	endIndex = startIndex + limit
	if endIndex > len(allRepos) {
		endIndex = len(allRepos)
	}

	paginatedRepos := allRepos[startIndex:endIndex]

	response := RepositoriesResponse{
		TotalCount:   len(allRepos),
		Repositories: paginatedRepos,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (handler *Handler) fetchAndAccumulateRepos(language, license, org string) ([]Repository, error) {
	var allRepos []Repository
	seen := make(map[string]struct{})
	if org != "" {
		page := 1
		for {
			repos, err := handler.fetchRepositoriesWithFilters("", language, license, org, page)
			if err != nil {
				return nil, err
			}
			if len(repos) == 0 {
				break
			}

			for _, repo := range repos {
				if _, exists := seen[repo.FullName]; !exists {
					if (language == "" || repo.Language == language) && (license == "" || (repo.License != nil && repo.License.Key == license)) {
						allRepos = append(allRepos, repo)
						seen[repo.FullName] = struct{}{}
					}
				}
			}
			page++
			if len(repos) == 100 {
				break
			}
		}
	} else {
		date := time.Now()
		for len(allRepos) < 100 {
			repos, err := handler.fetchRepositoriesWithFilters(date.Format("2006-01-02"), language, license, "", 1)
			if err != nil {
				return nil, err
			}
			for _, repo := range repos {
				if _, exists := seen[repo.FullName]; !exists {
					allRepos = append(allRepos, repo)
					seen[repo.FullName] = struct{}{}
				}
			}
			if len(repos) < 100 {
				date = date.AddDate(0, 0, -1)
			} else {
				break
			}
		}
	}
	return allRepos, nil
}
