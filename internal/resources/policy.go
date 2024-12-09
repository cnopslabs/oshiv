package resources

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/cnopslabs/oshiv/internal/utils"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

type Policy struct {
	name       string
	id         string
	statements []string
}

// Adding this because there's no set object type, may be worth implementing my own
// Checks if Policy object exists in Policy list by name
func policyContains(policies []Policy, policy Policy) bool {

	var policyExists bool

	for _, existing_policy := range policies {
		if policy.name == existing_policy.name {
			policyExists = true
		} else {
			policyExists = false
		}
	}

	return policyExists
}

// Fetch all policies via OCI API call
func fetchPolicies(identityClient identity.IdentityClient, compartmentId string) []Policy {
	var policies []Policy
	var pageCount int
	pageCount = 0

	initialResponse, err := identityClient.ListPolicies(context.Background(), identity.ListPoliciesRequest{CompartmentId: &compartmentId})
	utils.CheckError(err)

	for _, policy := range initialResponse.Items {
		pageCount += 1

		newPolicy := Policy{
			*policy.Name,
			*policy.Id,
			policy.Statements,
		}

		policies = append(policies, newPolicy)
	}

	if initialResponse.OpcNextPage != nil {
		pageCount += 1
		nextPage := initialResponse.OpcNextPage

		for {
			response, err := identityClient.ListPolicies(context.Background(), identity.ListPoliciesRequest{CompartmentId: &compartmentId, Page: nextPage})
			utils.CheckError(err)

			for _, policy := range response.Items {
				pageCount += 1

				newPolicy := Policy{
					*policy.Name,
					*policy.Id,
					policy.Statements,
				}

				policies = append(policies, newPolicy)
			}

			if response.OpcNextPage != nil {
				nextPage = response.OpcNextPage
			} else {
				break
			}
		}
	}

	return policies
}

// List and print policies (OCI API call)
func ListPolicies(identityClient identity.IdentityClient, compartmentId string, flagPolicyListNameOnly bool) {
	policies := fetchPolicies(identityClient, compartmentId)

	utils.Faint.Println(strconv.Itoa(len(policies)) + " results")

	for _, policy := range policies {
		if flagPolicyListNameOnly {
			fmt.Println(policy.name)
		} else {
			fmt.Print("Name: ")
			utils.Blue.Println(policy.name)

			fmt.Print("ID: ")
			utils.Yellow.Println(policy.id)

			fmt.Println("Statements: ")

			for _, statement := range policy.statements {
				utils.Faint.Println(statement)
			}
			fmt.Println("")
		}
	}
}

// Find and print policies (OCI API call)
func FindPolicies(identityClient identity.IdentityClient, compartmentId string, flagPolicyFind string, flagPolicyFindStatement string, flagPolicyListNameOnly bool) {
	// TODO: When matching on policy statement, it would probably make more sense to only return the statements with matches as opposed to returning all statements
	pattern_name := flagPolicyFind
	pattern_statement := flagPolicyFindStatement

	policies := fetchPolicies(identityClient, compartmentId)

	var matches []Policy

	// Handle simple wildcard
	if pattern_name == "*" {
		pattern_name = ".*"
	}

	if pattern_statement == "*" {
		pattern_statement = ".*"
	}

	if pattern_name != "" && pattern_statement == "" {
		// Match on name
		for _, policy := range policies {
			match, _ := regexp.MatchString(pattern_name, policy.name)
			if match {
				matches = append(matches, policy)
			}
		}
	} else if pattern_name == "" && pattern_statement != "" {
		// Match on statement
		for _, policy := range policies {
			for _, statement := range policy.statements {
				match, _ := regexp.MatchString(pattern_statement, statement)
				if match {
					if !policyContains(matches, policy) {
						matches = append(matches, policy)
					}
				}
			}
		}
	} else {
		// Match on policy name, then search only those policies for matches in statements
		var name_matches []Policy

		for _, policy := range policies {
			n_match, _ := regexp.MatchString(pattern_name, policy.name)
			if n_match {
				name_matches = append(name_matches, policy)
			}
		}

		for _, policy := range name_matches {
			for _, statement := range policy.statements {
				s_match, _ := regexp.MatchString(pattern_statement, statement)

				if s_match {
					if !policyContains(matches, policy) {
						matches = append(matches, policy)
					}
				}
			}
		}
	}

	if len(matches) > 0 {
		matchCount := len(matches)
		utils.Faint.Println(strconv.Itoa(matchCount) + " policy matches")

		for _, policy := range matches {
			if flagPolicyListNameOnly {
				// fmt.Print("Name: ")
				utils.Blue.Println(policy.name)
			} else {
				fmt.Print("Name: ")
				utils.Blue.Println(policy.name)

				fmt.Print("ID: ")
				utils.Yellow.Println(policy.id)

				fmt.Println("Statements: ")
				for _, statement := range policy.statements {
					utils.Faint.Println(statement)
				}

				fmt.Println("")
			}
		}
	}
}
