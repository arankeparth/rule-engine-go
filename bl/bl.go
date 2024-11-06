package bl

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"rule-engine/utils"
	"strings"
	"sync"
)

// Rule structure to define conditions and response
type Rule struct {
	Conditions map[string]map[string]string
	Response   string
}

// Load rules and cached responses
var (
	rules           []Rule
	responseCache   sync.Map
	rulesConfigPath = "rules.json" // Your rule config file path
)

// LoadRules loads rules from the configuration file
func LoadRules() error {
	data, err := ioutil.ReadFile(rulesConfigPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &rules); err != nil {
		return err
	}
	log.Printf("Rules loaded: %+v", rules)
	return nil
}

func MatchHeaders(headers map[string]string) []string {
	bestMatch := []string{}
	var maxScore int
	var mutex sync.Mutex

	// Check if we have a cached response for these headers.
	// This will optimize the performance in case the number of rules is very high.

	if cachedResponse, found := utils.GetValue(headers); found {
		return cachedResponse
	}

	var wg sync.WaitGroup
	processPart := func(part []Rule) {
		defer wg.Done()
		localMaxScore := 0
		localBestRules := []string{}
		for _, rule := range part {
			score := 0

			// Check 'equals' conditions
			if equalsConditions, ok := rule.Conditions[equalsKey]; ok {
				for key, expectedValue := range equalsConditions {
					if headers[key] == expectedValue {
						score++
					} else {
						score = 0
						break
					}
				}
			}

			// Check 'not_equals' conditions
			if notEqualsConditions, ok := rule.Conditions[notEqualsKey]; ok {
				for key, notExpectedValue := range notEqualsConditions {
					if headers[key] != notExpectedValue {
						score++
					} else {
						score = 0
						break
					}
				}
			}

			// Check 'contains' conditions
			if containsConditions, ok := rule.Conditions[containsKey]; ok {
				for key, substring := range containsConditions {
					if strings.Contains(headers[key], substring) {
						score++
					} else {
						score = 0
						break
					}
				}
			}

			// Update the local maximum score
			if score > localMaxScore {
				localMaxScore = score
				localBestRules = []string{rule.Response}
			} else if score == localMaxScore {
				localBestRules = append(localBestRules, rule.Response)
			}
		}

		// Update the global maxScore safely
		mutex.Lock()
		if localMaxScore > maxScore {
			maxScore = localMaxScore
			bestMatch = localBestRules
		} else if localMaxScore == maxScore {
			bestMatch = append(bestMatch, localBestRules...)
		}
		mutex.Unlock()
	}

	// Divide the rules into four parts
	partSize := (len(rules) + 3) / 4 // Ensure rounding up to cover all rules
	part1 := rules[:partSize]
	part2 := rules[partSize : 2*partSize]
	part3 := rules[2*partSize : 3*partSize]
	part4 := rules[3*partSize:]

	// Process each part concurrently
	wg.Add(4)
	go processPart(part1)
	go processPart(part2)
	go processPart(part3)
	go processPart(part4)

	// Wait for all parts to finish processing
	wg.Wait()

	if maxScore != 0 {
		// Cache the response before returning
		utils.StoreValue(headers, bestMatch)
		log.Printf("Best match: %+v", bestMatch)
		return bestMatch
	}
	return []string{noMatchRespFile}
}

// HandleRequest processes incoming requests
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	headers := make(map[string]string)
	for key, values := range r.Header {
		headers[key] = values[0]
	}

	// Find the best matching rule response
	responseFiles := MatchHeaders(headers)
	var jsonArray []map[string]interface{}

	// Iterate through response files
	for _, responseFile := range responseFiles {
		var jsonData map[string]interface{}

		// Check if the response is already cached
		if cachedResponse, found := responseCache.Load(responseFile); found {
			if err := json.Unmarshal(cachedResponse.([]byte), &jsonData); err != nil {
				http.Error(w, "Invalid cached JSON format", http.StatusInternalServerError)
				return
			}
			jsonArray = append(jsonArray, jsonData)
			continue
		}

		// Load the JSON response file from disk if not cached
		data, err := ioutil.ReadFile(responseFile)
		if err != nil {
			http.Error(w, "Response file not found", http.StatusNotFound)
			return
		}

		// Parse the JSON data
		if err := json.Unmarshal(data, &jsonData); err != nil {
			http.Error(w, "Invalid JSON format in file", http.StatusInternalServerError)
			return
		}

		// Cache the loaded response for future requests
		responseCache.Store(responseFile, data)
		jsonArray = append(jsonArray, jsonData)
	}

	// Convert the array of JSON objects into a JSON array
	responseData, err := json.Marshal(jsonArray)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	// Send the combined JSON array response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(responseData)
}
