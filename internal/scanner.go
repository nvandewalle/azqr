// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package internal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/renderers/csv"
	"github.com/Azure/azqr/internal/renderers/excel"
	"github.com/Azure/azqr/internal/renderers/json"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	_ "github.com/Azure/azqr/internal/scanners/aa"
	_ "github.com/Azure/azqr/internal/scanners/adf"
	_ "github.com/Azure/azqr/internal/scanners/afd"
	_ "github.com/Azure/azqr/internal/scanners/afw"
	_ "github.com/Azure/azqr/internal/scanners/agw"
	_ "github.com/Azure/azqr/internal/scanners/aks"
	_ "github.com/Azure/azqr/internal/scanners/amg"
	_ "github.com/Azure/azqr/internal/scanners/apim"
	_ "github.com/Azure/azqr/internal/scanners/appcs"
	_ "github.com/Azure/azqr/internal/scanners/appi"
	_ "github.com/Azure/azqr/internal/scanners/as"
	_ "github.com/Azure/azqr/internal/scanners/asp"
	_ "github.com/Azure/azqr/internal/scanners/avail"
	_ "github.com/Azure/azqr/internal/scanners/avd"
	_ "github.com/Azure/azqr/internal/scanners/avs"
	_ "github.com/Azure/azqr/internal/scanners/ba"
	_ "github.com/Azure/azqr/internal/scanners/ca"
	_ "github.com/Azure/azqr/internal/scanners/cae"
	_ "github.com/Azure/azqr/internal/scanners/ci"
	_ "github.com/Azure/azqr/internal/scanners/cog"
	_ "github.com/Azure/azqr/internal/scanners/conn"
	_ "github.com/Azure/azqr/internal/scanners/cosmos"
	_ "github.com/Azure/azqr/internal/scanners/cr"
	_ "github.com/Azure/azqr/internal/scanners/dbw"
	_ "github.com/Azure/azqr/internal/scanners/dec"
	_ "github.com/Azure/azqr/internal/scanners/disk"
	_ "github.com/Azure/azqr/internal/scanners/erc"
	_ "github.com/Azure/azqr/internal/scanners/evgd"
	_ "github.com/Azure/azqr/internal/scanners/evh"
	_ "github.com/Azure/azqr/internal/scanners/fdfp"
	_ "github.com/Azure/azqr/internal/scanners/gal"
	_ "github.com/Azure/azqr/internal/scanners/hpc"
	_ "github.com/Azure/azqr/internal/scanners/iot"
	_ "github.com/Azure/azqr/internal/scanners/it"
	_ "github.com/Azure/azqr/internal/scanners/kv"
	_ "github.com/Azure/azqr/internal/scanners/lb"
	_ "github.com/Azure/azqr/internal/scanners/log"
	_ "github.com/Azure/azqr/internal/scanners/logic"
	_ "github.com/Azure/azqr/internal/scanners/maria"
	_ "github.com/Azure/azqr/internal/scanners/mysql"
	_ "github.com/Azure/azqr/internal/scanners/netapp"
	_ "github.com/Azure/azqr/internal/scanners/ng"
	_ "github.com/Azure/azqr/internal/scanners/nic"
	_ "github.com/Azure/azqr/internal/scanners/nsg"
	_ "github.com/Azure/azqr/internal/scanners/nw"
	_ "github.com/Azure/azqr/internal/scanners/pdnsz"
	_ "github.com/Azure/azqr/internal/scanners/pep"
	_ "github.com/Azure/azqr/internal/scanners/pip"
	_ "github.com/Azure/azqr/internal/scanners/psql"
	_ "github.com/Azure/azqr/internal/scanners/redis"
	_ "github.com/Azure/azqr/internal/scanners/rg"
	_ "github.com/Azure/azqr/internal/scanners/rsv"
	_ "github.com/Azure/azqr/internal/scanners/rt"
	_ "github.com/Azure/azqr/internal/scanners/sap"
	_ "github.com/Azure/azqr/internal/scanners/sb"
	_ "github.com/Azure/azqr/internal/scanners/sigr"
	_ "github.com/Azure/azqr/internal/scanners/sql"
	_ "github.com/Azure/azqr/internal/scanners/st"
	_ "github.com/Azure/azqr/internal/scanners/synw"
	_ "github.com/Azure/azqr/internal/scanners/traf"
	_ "github.com/Azure/azqr/internal/scanners/vdpool"
	_ "github.com/Azure/azqr/internal/scanners/vgw"
	_ "github.com/Azure/azqr/internal/scanners/vm"
	_ "github.com/Azure/azqr/internal/scanners/vmss"
	_ "github.com/Azure/azqr/internal/scanners/vnet"
	_ "github.com/Azure/azqr/internal/scanners/vwan"
	_ "github.com/Azure/azqr/internal/scanners/wps"
)

type (
	ScanParams struct {
		ManagementGroupID       string
		SubscriptionID          string
		ResourceGroup           string
		OutputName              string
		Defender                bool
		Advisor                 bool
		Cost                    bool
		Mask                    bool
		Csv                     bool
		Json                    bool
		Debug                   bool
		ScannerKeys             []string
		ForceAzureCliCredential bool
		Filters                 *models.Filters
		UseAzqrRecommendations  bool
		UseAprlRecommendations  bool
	}

	Scanner struct{}
)

func NewScanParams() *ScanParams {
	return &ScanParams{
		SubscriptionID:          "",
		ResourceGroup:           "",
		OutputName:              "",
		Defender:                true,
		Advisor:                 true,
		Cost:                    true,
		Mask:                    true,
		Csv:                     false,
		Json:                    false,
		Debug:                   false,
		ScannerKeys:             []string{},
		ForceAzureCliCredential: false,
		Filters:                 models.NewFilters(),
		UseAzqrRecommendations:  true,
		UseAprlRecommendations:  true,
	}
}

func (sc Scanner) Scan(params *ScanParams) {
	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if params.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled")
	}

	// generate output file name
	outputFile := sc.generateOutputFileName(params.OutputName)

	// load filters
	filters := params.Filters

	// validate input
	if params.ManagementGroupID != "" && (params.SubscriptionID != "" || params.ResourceGroup != "") {
		log.Fatal().Msg("Management Group name cannot be used with a Subscription Id or Resource Group name")
	}

	if params.SubscriptionID == "" && params.ResourceGroup != "" {
		log.Fatal().Msg("Resource Group name can only be used with a Subscription Id")
	}

	if params.SubscriptionID != "" {
		filters.Azqr.AddSubscription(params.SubscriptionID)
	}

	if params.ResourceGroup != "" {
		filters.Azqr.AddResourceGroup(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", params.SubscriptionID, params.ResourceGroup))
	}

	serviceScanners := filters.Azqr.Scanners

	// create Azure credentials
	cred := sc.newAzureCredential(params.ForceAzureCliCredential)

	// create a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create ARM client options
	clientOptions := &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Retry: policy.RetryOptions{
				RetryDelay:    20 * time.Millisecond,
				MaxRetries:    3,
				MaxRetryDelay: 10 * time.Minute,
			},
		},
	}

	// list subscriptions. Key is subscription ID, value is subscription name
	var subscriptions map[string]string
	if params.ManagementGroupID != "" {
		managementGroupScanner := scanners.ManagementGroupsScanner{}
		subscriptions = managementGroupScanner.ListSubscriptions(ctx, cred, params.ManagementGroupID, filters, clientOptions)
	} else {
		subscriptionScanner := scanners.SubcriptionScanner{}
		subscriptions = subscriptionScanner.ListSubscriptions(ctx, cred, params.SubscriptionID, filters, clientOptions)
	}

	// initialize scanners
	defenderScanner := scanners.DefenderScanner{}
	pipScanner := scanners.PublicIPScanner{}
	peScanner := scanners.PrivateEndpointScanner{}
	diagnosticsScanner := scanners.DiagnosticSettingsScanner{}
	advisorScanner := scanners.AdvisorScanner{}
	costScanner := scanners.CostScanner{}
	diagResults := map[string]bool{}

	// initialize report data
	reportData := renderers.NewReportData(outputFile, params.Mask)

	// get the APRL scan results
	aprlScanner := graph.NewAprlScanner(serviceScanners, filters, subscriptions)
	reportData.Recommendations, reportData.Aprl = aprlScanner.Scan(ctx, cred)

	resourceScanner := scanners.ResourceScanner{}
	reportData.Resources, reportData.ExludedResources = resourceScanner.GetAllResources(ctx, cred, subscriptions, filters)

	// For each service scanner, get the recommendations list
	if params.UseAzqrRecommendations {
		for _, s := range serviceScanners {
			for i, r := range s.GetRecommendations() {
				if filters.Azqr.IsRecommendationExcluded(r.RecommendationID) {
					continue
				}

				if r.RecommendationType != models.TypeRecommendation {
					continue
				}

				if reportData.Recommendations[strings.ToLower(r.ResourceType)] == nil {
					reportData.Recommendations[strings.ToLower(r.ResourceType)] = map[string]models.AprlRecommendation{}
				}

				reportData.Recommendations[strings.ToLower(r.ResourceType)][i] = r.ToAzureAprlRecommendation()
			}
		}

		// scan diagnostic settings
		err := diagnosticsScanner.Init(ctx, cred, clientOptions)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize diagnostic settings scanner")
		}

		diagResults = diagnosticsScanner.Scan(reportData.ResourceIDs())
	}

	// scan each subscription with AZQR scanners
	for sid, sn := range subscriptions {
		config := &models.ScannerConfig{
			Ctx:              ctx,
			SubscriptionID:   sid,
			SubscriptionName: sn,
			Cred:             cred,
			ClientOptions:    clientOptions,
		}

		if params.UseAzqrRecommendations {
			// scan private endpoints
			peResults := peScanner.Scan(config)

			// scan public IPs
			pips := pipScanner.Scan(config)

			// initialize scan context
			scanContext := models.ScanContext{
				Filters:             filters,
				PrivateEndpoints:    peResults,
				DiagnosticsSettings: diagResults,
				PublicIPs:           pips,
			}

			// scan each resource group
			ch := make(chan []models.AzqrServiceResult, len(serviceScanners))

			for _, s := range serviceScanners {
				err := s.Init(config)
				if err != nil {
					log.Fatal().Err(err).Msg("Failed to initialize scanner")
				}

				go func(s models.IAzureScanner) {
					res, err := sc.retry(3, 10*time.Millisecond, s, &scanContext)
					if err != nil {
						cancel()
						log.Fatal().Err(err).Msg("Failed to scan")
					}
					ch <- res
				}(s)
			}

			for i := 0; i < len(serviceScanners); i++ {
				res := <-ch
				for _, r := range res {
					// check if the resource is excluded
					if filters.Azqr.IsServiceExcluded(r.ResourceID()) {
						continue
					}
					reportData.Azqr = append(reportData.Azqr, r)
				}
			}
		}

		// scan costs
		costs := costScanner.Scan(params.Cost, config)
		reportData.Cost.From = costs.From
		reportData.Cost.To = costs.To
		reportData.Cost.Items = append(reportData.Cost.Items, costs.Items...)
	}

	// get the count of resources per resource type
	reportData.ResourceTypeCount = resourceScanner.GetCountPerResourceType(ctx, cred, subscriptions, reportData.Recommendations, filters)

	// scan advisor
	reportData.Advisor = append(reportData.Advisor, advisorScanner.Scan(ctx, params.Defender, cred, subscriptions, filters)...)

	// scan defender
	reportData.Defender = append(reportData.Defender, defenderScanner.Scan(ctx, params.Defender, cred, subscriptions, filters)...)

	// get the defender recommendations
	reportData.DefenderRecommendations = append(reportData.DefenderRecommendations, defenderScanner.GetRecommendations(ctx, params.Defender, cred, subscriptions, filters)...)

	// render excel report
	excel.CreateExcelReport(&reportData)

	// render json report
	if params.Json {
		json.CreateJsonReport(&reportData)
	}

	// render csv reports
	if params.Csv {
		csv.CreateCsvReport(&reportData)
	}

	log.Info().Msg("Scan completed.")
}

// retry retries the Azure scanner Scan, a number of times with an increasing delay between retries
func (sc Scanner) retry(attempts int, sleep time.Duration, a models.IAzureScanner, scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	var err error
	for i := 0; ; i++ {
		res, err := a.Scan(scanContext)
		if err == nil {
			return res, nil
		}

		if models.ShouldSkipError(err) {
			return []models.AzqrServiceResult{}, nil
		}

		errAsString := err.Error()

		if i >= (attempts - 1) {
			log.Info().Msgf("Retry limit reached. Error: %s", errAsString)
			break
		}

		log.Debug().Msgf("Retrying after error: %s", errAsString)

		time.Sleep(sleep)
		sleep *= 2
	}
	return nil, err
}

func (sc Scanner) newAzureCredential(forceAzureCliCredential bool) azcore.TokenCredential {
	var cred azcore.TokenCredential
	var err error
	if !forceAzureCliCredential {
		cred, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Azure credentials")
		}
	} else {
		cred, err = azidentity.NewAzureCLICredential(nil)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Azure CLI credentials")
		}
	}
	return cred
}

func (sc Scanner) generateOutputFileName(outputName string) string {
	outputFile := outputName
	if outputFile == "" {
		current_time := time.Now()
		outputFileStamp := fmt.Sprintf("%d_%02d_%02d_T%02d%02d%02d",
			current_time.Year(), current_time.Month(), current_time.Day(),
			current_time.Hour(), current_time.Minute(), current_time.Second())

		outputFile = fmt.Sprintf("%s_%s", "azqr_action_plan", outputFileStamp)
	}
	return outputFile
}
