package analytics

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/rusiqe/domainvault/internal/storage"
	"github.com/rusiqe/domainvault/internal/types"
)

// AnalyticsService provides domain portfolio analytics
type AnalyticsService struct {
	domainRepo storage.DomainRepository
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(domainRepo storage.DomainRepository) *AnalyticsService {
	return &AnalyticsService{
		domainRepo: domainRepo,
	}
}

// PortfolioMetrics represents comprehensive portfolio analytics
type PortfolioMetrics struct {
	Overview            OverviewMetrics            `json:"overview"`
	FinancialMetrics    FinancialMetrics          `json:"financial_metrics"`
	ExpirationAnalysis  ExpirationAnalysis        `json:"expiration_analysis"`
	ProviderAnalysis    ProviderAnalysis          `json:"provider_analysis"`
	CategoryAnalysis    CategoryAnalysis          `json:"category_analysis"`
	StatusMetrics       StatusMetrics             `json:"status_metrics"`
	TrendAnalysis       TrendAnalysis             `json:"trend_analysis"`
	RiskAssessment      RiskAssessment            `json:"risk_assessment"`
	Recommendations     []Recommendation          `json:"recommendations"`
	LastUpdated         time.Time                 `json:"last_updated"`
}

// OverviewMetrics provides high-level portfolio overview
type OverviewMetrics struct {
	TotalDomains        int       `json:"total_domains"`
	ActiveDomains       int       `json:"active_domains"`
	ExpiredDomains      int       `json:"expired_domains"`
	DomainsExpiring30   int       `json:"domains_expiring_30"`
	DomainsExpiring7    int       `json:"domains_expiring_7"`
	AverageAge          float64   `json:"average_age_days"`
	OldestDomain        string    `json:"oldest_domain"`
	NewestDomain        string    `json:"newest_domain"`
	LastSyncTime        time.Time `json:"last_sync_time"`
}

// FinancialMetrics provides cost and value analysis
type FinancialMetrics struct {
	TotalRenewalCost       float64                      `json:"total_renewal_cost"`
	RenewalCostNext30Days  float64                      `json:"renewal_cost_next_30_days"`
	RenewalCostNext90Days  float64                      `json:"renewal_cost_next_90_days"`
	AverageRenewalCost     float64                      `json:"average_renewal_cost"`
	CostByProvider         map[string]float64           `json:"cost_by_provider"`
	CostByCategory         map[string]float64           `json:"cost_by_category"`
	MonthlyRenewalSchedule map[string]float64           `json:"monthly_renewal_schedule"`
	EstimatedValue         EstimatedValueMetrics        `json:"estimated_value"`
}

// EstimatedValueMetrics represents domain portfolio valuation
type EstimatedValueMetrics struct {
	TotalEstimatedValue    float64            `json:"total_estimated_value"`
	AverageValuePerDomain  float64            `json:"average_value_per_domain"`
	ValueByCategory        map[string]float64 `json:"value_by_category"`
	PremiumDomains         []DomainValue      `json:"premium_domains"`
	ValueDistribution      ValueDistribution  `json:"value_distribution"`
}

// DomainValue represents a domain's estimated value
type DomainValue struct {
	DomainName       string  `json:"domain_name"`
	EstimatedValue   float64 `json:"estimated_value"`
	RenewalCost      float64 `json:"renewal_cost"`
	ValueMultiplier  float64 `json:"value_multiplier"`
	ValuationFactors []string `json:"valuation_factors"`
}

// ValueDistribution represents value ranges
type ValueDistribution struct {
	Under100      int `json:"under_100"`
	Between100500 int `json:"between_100_500"`
	Between5001K  int `json:"between_500_1k"`
	Between1K5K   int `json:"between_1k_5k"`
	Above5K       int `json:"above_5k"`
}

// ExpirationAnalysis provides expiration insights
type ExpirationAnalysis struct {
	ExpirationDistribution map[string]int            `json:"expiration_distribution"`
	CriticalDomains        []ExpiringDomain          `json:"critical_domains"`
	MonthlyExpirations     map[string]int            `json:"monthly_expirations"`
	AutoRenewStatus        AutoRenewStats            `json:"auto_renew_status"`
	ExpirationHeatmap      map[string]int            `json:"expiration_heatmap"`
}

// ExpiringDomain represents a domain approaching expiration
type ExpiringDomain struct {
	DomainName      string    `json:"domain_name"`
	ExpiresAt       time.Time `json:"expires_at"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
	RenewalCost     float64   `json:"renewal_cost"`
	AutoRenew       bool      `json:"auto_renew"`
	Provider        string    `json:"provider"`
	RiskLevel       string    `json:"risk_level"`
}

// AutoRenewStats represents auto-renewal statistics
type AutoRenewStats struct {
	Enabled          int     `json:"enabled"`
	Disabled         int     `json:"disabled"`
	PercentageEnabled float64 `json:"percentage_enabled"`
}

// ProviderAnalysis provides registrar performance insights
type ProviderAnalysis struct {
	ProviderDistribution   map[string]int             `json:"provider_distribution"`
	ProviderReliability    map[string]float64         `json:"provider_reliability"`
	ProviderCostEfficiency map[string]float64         `json:"provider_cost_efficiency"`
	ProviderPerformance    map[string]ProviderMetrics `json:"provider_performance"`
	RecommendedProviders   []string                   `json:"recommended_providers"`
}

// ProviderMetrics represents performance metrics for a provider
type ProviderMetrics struct {
	DomainCount         int     `json:"domain_count"`
	AverageRenewalCost  float64 `json:"average_renewal_cost"`
	UptimePercentage    float64 `json:"uptime_percentage"`
	SyncReliability     float64 `json:"sync_reliability"`
	LastSyncError       string  `json:"last_sync_error,omitempty"`
	RecommendationScore float64 `json:"recommendation_score"`
}

// CategoryAnalysis provides domain categorization insights
type CategoryAnalysis struct {
	CategoryDistribution map[string]int    `json:"category_distribution"`
	CategoryValue        map[string]float64 `json:"category_value"`
	CategoryPerformance  map[string]float64 `json:"category_performance"`
	UncategorizedDomains int               `json:"uncategorized_domains"`
}

// StatusMetrics provides health monitoring insights
type StatusMetrics struct {
	StatusDistribution   map[string]int             `json:"status_distribution"`
	UptimeStats          UptimeStats               `json:"uptime_stats"`
	ResponseTimeMetrics  ResponseTimeMetrics       `json:"response_time_metrics"`
	StatusTrends         map[string][]StatusTrend  `json:"status_trends"`
	ProblematicDomains   []ProblematicDomain       `json:"problematic_domains"`
}

// UptimeStats represents overall uptime statistics
type UptimeStats struct {
	OverallUptime     float64 `json:"overall_uptime"`
	DomainsOnline     int     `json:"domains_online"`
	DomainsOffline    int     `json:"domains_offline"`
	DomainsUnknown    int     `json:"domains_unknown"`
	LastCheckTime     time.Time `json:"last_check_time"`
}

// ResponseTimeMetrics represents response time analysis
type ResponseTimeMetrics struct {
	AverageResponseTime float64   `json:"average_response_time"`
	MedianResponseTime  float64   `json:"median_response_time"`
	P95ResponseTime     float64   `json:"p95_response_time"`
	FastestDomain       string    `json:"fastest_domain"`
	SlowestDomain       string    `json:"slowest_domain"`
}

// StatusTrend represents status changes over time
type StatusTrend struct {
	Date   time.Time `json:"date"`
	Status int       `json:"status"`
	Count  int       `json:"count"`
}

// ProblematicDomain represents domains with issues
type ProblematicDomain struct {
	DomainName       string    `json:"domain_name"`
	Issue            string    `json:"issue"`
	LastGoodStatus   int       `json:"last_good_status"`
	DowntimeDuration string    `json:"downtime_duration"`
	ImpactLevel      string    `json:"impact_level"`
}

// TrendAnalysis provides historical trend insights
type TrendAnalysis struct {
	DomainGrowth        []GrowthTrend      `json:"domain_growth"`
	CostTrends          []CostTrend        `json:"cost_trends"`
	ExpirationTrends    []ExpirationTrend  `json:"expiration_trends"`
	StatusTrends        []StatusTrend      `json:"status_trends"`
	Seasonality         SeasonalityAnalysis `json:"seasonality"`
}

// GrowthTrend represents domain portfolio growth
type GrowthTrend struct {
	Period       string    `json:"period"`
	DomainCount  int       `json:"domain_count"`
	NetChange    int       `json:"net_change"`
	GrowthRate   float64   `json:"growth_rate"`
}

// CostTrend represents cost trends over time
type CostTrend struct {
	Period      string  `json:"period"`
	TotalCost   float64 `json:"total_cost"`
	AverageCost float64 `json:"average_cost"`
	CostChange  float64 `json:"cost_change"`
}

// ExpirationTrend represents expiration patterns
type ExpirationTrend struct {
	Period            string `json:"period"`
	ExpiredDomains    int    `json:"expired_domains"`
	RenewedDomains    int    `json:"renewed_domains"`
	RenewalRate       float64 `json:"renewal_rate"`
}

// SeasonalityAnalysis represents seasonal patterns
type SeasonalityAnalysis struct {
	PeakExpirationMonths []string `json:"peak_expiration_months"`
	CostSeasonality      map[string]float64 `json:"cost_seasonality"`
	StatusSeasonality    map[string]float64 `json:"status_seasonality"`
}

// RiskAssessment provides portfolio risk analysis
type RiskAssessment struct {
	OverallRiskScore      float64                  `json:"overall_risk_score"`
	RiskFactors           []RiskFactor             `json:"risk_factors"`
	HighRiskDomains       []HighRiskDomain         `json:"high_risk_domains"`
	RiskMitigation        []RiskMitigation         `json:"risk_mitigation"`
	SecurityMetrics       SecurityMetrics          `json:"security_metrics"`
	ComplianceStatus      ComplianceStatus         `json:"compliance_status"`
}

// RiskFactor represents identified risks
type RiskFactor struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	Impact      float64 `json:"impact"`
	Probability float64 `json:"probability"`
}

// HighRiskDomain represents domains with elevated risk
type HighRiskDomain struct {
	DomainName    string   `json:"domain_name"`
	RiskScore     float64  `json:"risk_score"`
	RiskReasons   []string `json:"risk_reasons"`
	Mitigation    []string `json:"mitigation"`
}

// RiskMitigation represents risk reduction strategies
type RiskMitigation struct {
	Strategy     string  `json:"strategy"`
	Priority     string  `json:"priority"`
	Effort       string  `json:"effort"`
	RiskReduction float64 `json:"risk_reduction"`
}

// SecurityMetrics represents security-related metrics
type SecurityMetrics struct {
	SSLCertificateStatus map[string]int `json:"ssl_certificate_status"`
	DNSSECStatus         map[string]int `json:"dnssec_status"`
	SecurityHeaders      map[string]int `json:"security_headers"`
	VulnerabilityCount   int            `json:"vulnerability_count"`
}

// ComplianceStatus represents compliance metrics
type ComplianceStatus struct {
	GDPRCompliant     int     `json:"gdpr_compliant"`
	PrivacyPolicy     int     `json:"privacy_policy"`
	TermsOfService    int     `json:"terms_of_service"`
	ComplianceScore   float64 `json:"compliance_score"`
}

// Recommendation represents actionable insights
type Recommendation struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Priority    string    `json:"priority"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	Effort      string    `json:"effort"`
	Actions     []string  `json:"actions"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetPortfolioMetrics generates comprehensive portfolio analytics
func (as *AnalyticsService) GetPortfolioMetrics() (*PortfolioMetrics, error) {
	domains, err := as.domainRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch domains: %w", err)
	}

	metrics := &PortfolioMetrics{
		Overview:           as.calculateOverviewMetrics(domains),
		FinancialMetrics:   as.calculateFinancialMetrics(domains),
		ExpirationAnalysis: as.calculateExpirationAnalysis(domains),
		ProviderAnalysis:   as.calculateProviderAnalysis(domains),
		CategoryAnalysis:   as.calculateCategoryAnalysis(domains),
		StatusMetrics:      as.calculateStatusMetrics(domains),
		TrendAnalysis:      as.calculateTrendAnalysis(domains),
		RiskAssessment:     as.calculateRiskAssessment(domains),
		Recommendations:    as.generateRecommendations(domains),
		LastUpdated:        time.Now(),
	}

	return metrics, nil
}

// calculateOverviewMetrics calculates basic portfolio overview
func (as *AnalyticsService) calculateOverviewMetrics(domains []types.Domain) OverviewMetrics {
	now := time.Now()
	activeDomains := 0
	expiredDomains := 0
	expiring30 := 0
	expiring7 := 0
	totalAge := 0.0
	oldestDomain := ""
	newestDomain := ""
	oldestDate := now
	newestDate := time.Time{}

	for _, domain := range domains {
		age := now.Sub(domain.CreatedAt).Hours() / 24
		totalAge += age

		if domain.CreatedAt.Before(oldestDate) {
			oldestDate = domain.CreatedAt
			oldestDomain = domain.Name
		}

		if domain.CreatedAt.After(newestDate) {
			newestDate = domain.CreatedAt
			newestDomain = domain.Name
		}

		daysUntilExpiry := int(domain.ExpiresAt.Sub(now).Hours() / 24)

		if daysUntilExpiry < 0 {
			expiredDomains++
		} else {
			activeDomains++
			if daysUntilExpiry <= 30 {
				expiring30++
			}
			if daysUntilExpiry <= 7 {
				expiring7++
			}
		}
	}

	averageAge := 0.0
	if len(domains) > 0 {
		averageAge = totalAge / float64(len(domains))
	}

	return OverviewMetrics{
		TotalDomains:        len(domains),
		ActiveDomains:       activeDomains,
		ExpiredDomains:      expiredDomains,
		DomainsExpiring30:   expiring30,
		DomainsExpiring7:    expiring7,
		AverageAge:          averageAge,
		OldestDomain:        oldestDomain,
		NewestDomain:        newestDomain,
		LastSyncTime:        now,
	}
}

// calculateFinancialMetrics calculates cost and value metrics
func (as *AnalyticsService) calculateFinancialMetrics(domains []types.Domain) FinancialMetrics {
	now := time.Now()
	totalRenewalCost := 0.0
	renewal30Days := 0.0
	renewal90Days := 0.0
	costByProvider := make(map[string]float64)
	costByCategory := make(map[string]float64)
	monthlySchedule := make(map[string]float64)

	validCosts := 0

	for _, domain := range domains {
		cost := 0.0
		if domain.RenewalPrice != nil {
			cost = *domain.RenewalPrice
			validCosts++
		} else {
			cost = as.estimateRenewalCost(domain)
		}

		totalRenewalCost += cost
		
		// Provider costs
		costByProvider[domain.Provider] += cost

		// Category costs
		category := "Uncategorized"
		if domain.CategoryID != nil {
			category = *domain.CategoryID
		}
		costByCategory[category] += cost

		// Renewal schedule
		daysUntilExpiry := int(domain.ExpiresAt.Sub(now).Hours() / 24)
		if daysUntilExpiry <= 30 && daysUntilExpiry > 0 {
			renewal30Days += cost
		}
		if daysUntilExpiry <= 90 && daysUntilExpiry > 0 {
			renewal90Days += cost
		}

		// Monthly schedule
		month := domain.ExpiresAt.Format("2006-01")
		monthlySchedule[month] += cost
	}

	averageRenewalCost := 0.0
	if len(domains) > 0 {
		averageRenewalCost = totalRenewalCost / float64(len(domains))
	}

	return FinancialMetrics{
		TotalRenewalCost:       totalRenewalCost,
		RenewalCostNext30Days:  renewal30Days,
		RenewalCostNext90Days:  renewal90Days,
		AverageRenewalCost:     averageRenewalCost,
		CostByProvider:         costByProvider,
		CostByCategory:         costByCategory,
		MonthlyRenewalSchedule: monthlySchedule,
		EstimatedValue:         as.calculateEstimatedValue(domains),
	}
}

// calculateEstimatedValue estimates domain portfolio value
func (as *AnalyticsService) calculateEstimatedValue(domains []types.Domain) EstimatedValueMetrics {
	totalValue := 0.0
	valueByCategory := make(map[string]float64)
	var premiumDomains []DomainValue
	distribution := ValueDistribution{}

	for _, domain := range domains {
		value := as.estimateDomainValue(domain)
		totalValue += value

		category := "Uncategorized"
		if domain.CategoryID != nil {
			category = *domain.CategoryID
		}
		valueByCategory[category] += value

		// Check if premium (value > 5x renewal cost)
		renewalCost := 15.0 // Default
		if domain.RenewalPrice != nil {
			renewalCost = *domain.RenewalPrice
		}

		if value > renewalCost*5 {
			premiumDomains = append(premiumDomains, DomainValue{
				DomainName:       domain.Name,
				EstimatedValue:   value,
				RenewalCost:      renewalCost,
				ValueMultiplier:  value / renewalCost,
				ValuationFactors: as.getValuationFactors(domain),
			})
		}

		// Value distribution
		switch {
		case value < 100:
			distribution.Under100++
		case value < 500:
			distribution.Between100500++
		case value < 1000:
			distribution.Between5001K++
		case value < 5000:
			distribution.Between1K5K++
		default:
			distribution.Above5K++
		}
	}

	// Sort premium domains by value
	sort.Slice(premiumDomains, func(i, j int) bool {
		return premiumDomains[i].EstimatedValue > premiumDomains[j].EstimatedValue
	})

	averageValue := 0.0
	if len(domains) > 0 {
		averageValue = totalValue / float64(len(domains))
	}

	return EstimatedValueMetrics{
		TotalEstimatedValue:   totalValue,
		AverageValuePerDomain: averageValue,
		ValueByCategory:       valueByCategory,
		PremiumDomains:        premiumDomains,
		ValueDistribution:     distribution,
	}
}

// Helper methods for value estimation
func (as *AnalyticsService) estimateRenewalCost(domain types.Domain) float64 {
	// Simple cost estimation based on provider and TLD
	baseCost := 15.0
	
	switch domain.Provider {
	case "namecheap":
		baseCost = 12.0
	case "godaddy":
		baseCost = 15.0
	case "cloudflare":
		baseCost = 9.0
	}

	// TLD adjustments
	if strings.Contains(domain.Name, ".io") {
		baseCost = 60.0
	} else if strings.Contains(domain.Name, ".ai") {
		baseCost = 200.0
	} else if strings.Contains(domain.Name, ".dev") {
		baseCost = 30.0
	}

	return baseCost
}

func (as *AnalyticsService) estimateDomainValue(domain types.Domain) float64 {
	baseValue := 50.0 // Base domain value

	// Length factor (shorter is better)
	nameLen := len(strings.Split(domain.Name, ".")[0])
	if nameLen <= 3 {
		baseValue *= 10
	} else if nameLen <= 5 {
		baseValue *= 5
	} else if nameLen <= 8 {
		baseValue *= 2
	}

	// TLD factor
	if strings.HasSuffix(domain.Name, ".com") {
		baseValue *= 3
	} else if strings.HasSuffix(domain.Name, ".io") {
		baseValue *= 2
	} else if strings.HasSuffix(domain.Name, ".ai") {
		baseValue *= 4
	}

	// Age factor
	age := time.Since(domain.CreatedAt).Hours() / 24 / 365
	if age > 5 {
		baseValue *= 1.5
	}

	// Status factor
	if domain.HTTPStatus != nil && *domain.HTTPStatus == 200 {
		baseValue *= 1.2
	}

	return math.Round(baseValue)
}

func (as *AnalyticsService) getValuationFactors(domain types.Domain) []string {
	var factors []string

	nameLen := len(strings.Split(domain.Name, ".")[0])
	if nameLen <= 5 {
		factors = append(factors, "Short name")
	}

	if strings.HasSuffix(domain.Name, ".com") {
		factors = append(factors, "Premium TLD (.com)")
	}

	age := time.Since(domain.CreatedAt).Hours() / 24 / 365
	if age > 5 {
		factors = append(factors, "Established domain (5+ years)")
	}

	if domain.HTTPStatus != nil && *domain.HTTPStatus == 200 {
		factors = append(factors, "Active website")
	}

	return factors
}

// Placeholder methods for other calculations
func (as *AnalyticsService) calculateExpirationAnalysis(domains []types.Domain) ExpirationAnalysis {
	// Implementation would calculate expiration patterns, critical domains, etc.
	return ExpirationAnalysis{}
}

func (as *AnalyticsService) calculateProviderAnalysis(domains []types.Domain) ProviderAnalysis {
	// Implementation would analyze provider performance, reliability, etc.
	return ProviderAnalysis{}
}

func (as *AnalyticsService) calculateCategoryAnalysis(domains []types.Domain) CategoryAnalysis {
	// Implementation would analyze category distribution and performance
	return CategoryAnalysis{}
}

func (as *AnalyticsService) calculateStatusMetrics(domains []types.Domain) StatusMetrics {
	// Implementation would analyze uptime, response times, status trends
	return StatusMetrics{}
}

func (as *AnalyticsService) calculateTrendAnalysis(domains []types.Domain) TrendAnalysis {
	// Implementation would analyze historical trends
	return TrendAnalysis{}
}

func (as *AnalyticsService) calculateRiskAssessment(domains []types.Domain) RiskAssessment {
	// Implementation would assess portfolio risks
	return RiskAssessment{}
}

func (as *AnalyticsService) generateRecommendations(domains []types.Domain) []Recommendation {
	// Implementation would generate actionable recommendations
	return []Recommendation{}
}
