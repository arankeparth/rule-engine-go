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

func MatchHeaders(headers map[string]string) string {
	var bestMatch Rule
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
		var localBestRule Rule
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
				localBestRule = rule
			}
		}

		// Update the global maxScore safely
		mutex.Lock()
		if localMaxScore > maxScore {
			maxScore = localMaxScore
			bestMatch = localBestRule
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
		utils.StoreValue(headers, bestMatch.Response)
		log.Printf("Best match: %+v", bestMatch)
		return bestMatch.Response
	}
	return noMatchRespFile
}

// HandleRequest processes incoming requests
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	headers := make(map[string]string)
	for key, values := range r.Header {
		headers[key] = values[0]
	}

	// Find the best matching rule response
	responseFile := MatchHeaders(headers)
	// // Load and cache the JSON response
	if cachedResponse, found := responseCache.Load(responseFile); found {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(cachedResponse.([]byte))
		w.WriteHeader(http.StatusOK)
		return
	}

	// Load JSON response file from disk and cache it
	data, err := ioutil.ReadFile(responseFile)
	if err != nil {
		http.Error(w, "Response file not found", http.StatusNotFound)
		return
	}

	// Cache the loaded response for future requests
	responseCache.Store(responseFile, data)

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
	w.WriteHeader(http.StatusOK)
}
