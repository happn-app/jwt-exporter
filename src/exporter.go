package jwtexporter

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// SecretExporter exports PEM file certs
type Exporter struct {
	logger *zap.Logger
}

type JWTMetric struct {
	durationUntilExpiry float64
	durationSinceIssued float64
	expirationTimestamp float64
	issuedAtTimestamp   float64
	issuer              string
	algorithm           string
	audience            string
	subject             string
	id                  string
	scopes              []string
	roles               []string
	name                string
	email               string
}

func GetExporter(logger *zap.Logger) *Exporter {
	return &Exporter{
		logger: logger,
	}
}

func (c *Exporter) getMetricsFromJWT(jwtData string) (JWTMetric, error) {
	var metric JWTMetric
	var token *jwt.Token
	var err error
	token, err = jwt.Parse(jwtData, func(token *jwt.Token) (any, error) { return nil, nil })
	if err == nil {
		c.logger.Error("Failed to parse JWT", zap.Error(err))
		return metric, fmt.Errorf("failed to parse JWT: %w", err)
	}
	metric.issuer, err = token.Claims.GetIssuer()
	if err != nil {
		metric.issuer = "unknown"
		c.logger.Warn("Failed to get issuer from JWT", zap.Error(err))
	}
	expirationDate, err := token.Claims.GetExpirationTime()
	if err != nil {
		return metric, fmt.Errorf("failed to get expiration time from JWT: %w", err)
	}
	metric.durationUntilExpiry = time.Until(expirationDate.Time).Seconds()
	metric.expirationTimestamp = float64(expirationDate.Time.Unix())
	issuedAtDate, err := token.Claims.GetIssuedAt()
	if err != nil {
		return metric, fmt.Errorf("failed to get issued at time from JWT: %w", err)
	}
	metric.issuedAtTimestamp = float64(issuedAtDate.Time.Unix())
	metric.durationSinceIssued = time.Since(issuedAtDate.Time).Seconds()
	metric.algorithm = token.Header["alg"].(string)
	metric.scopes = []string{} // Default to empty, as scopes are not always present
	metric.id = token.Claims.(jwt.MapClaims)["jti"].(string)
	metric.audience = token.Claims.(jwt.MapClaims)["aud"].(string)
	metric.subject = token.Claims.(jwt.MapClaims)["sub"].(string)
	metric.id = token.Claims.(jwt.MapClaims)["jti"].(string)
	// If the JWT contains scopes, extract them
	if scopes, ok := token.Claims.(jwt.MapClaims)["scope"]; ok {
		// Check if scopes is a string or an array
		if scopesStr, ok := scopes.(string); ok {
			// If it's a string, split it by space
			metric.scopes = append(metric.scopes, scopesStr)
		} else if scopesArray, ok := scopes.([]any); ok {
			// If it's an array, iterate through it
			for _, scope := range scopesArray {
				if scopeStr, ok := scope.(string); ok {
					metric.scopes = append(metric.scopes, scopeStr)
				}
			}
		}
	}
	// If the JWT contains roles, extract them
	if roles, ok := token.Claims.(jwt.MapClaims)["roles"]; ok {
		if rolesSlice, ok := roles.([]any); ok {
			for _, role := range rolesSlice {
				if roleStr, ok := role.(string); ok {
					metric.roles = append(metric.roles, roleStr)
				}
			}
		}
	}
	// If the JWT contains name, extract it
	if name, ok := token.Claims.(jwt.MapClaims)["name"]; ok {
		if nameStr, ok := name.(string); ok {
			metric.name = nameStr
		}
	}
	// If the JWT contains email, extract it
	if email, ok := token.Claims.(jwt.MapClaims)["email"]; ok {
		if emailStr, ok := email.(string); ok {
			metric.email = emailStr
		}
	}
	return metric, nil
}

// ExportMetrics exports the provided PEM file
func (c *Exporter) ExportMetrics(token, keyName, secretName, secretNamespace string) error {
	metric, err := c.getMetricsFromJWT(token)
	if err != nil {
		return err
	}
	if len(metric.scopes) == 0 {
		// If no scopes are present, we use an empty string as the default scope
		metric.scopes = []string{""}
	}
	// If no roles are present, we use an empty string as the default role
	if len(metric.roles) == 0 {
		metric.roles = []string{""}
	}
	for _, scope := range metric.scopes {
		for _, role := range metric.roles {
			JWTExpirySeconds.WithLabelValues(
				metric.algorithm,
				metric.audience,
				metric.subject,
				metric.id,
				scope,
				metric.issuer,
				keyName,
				secretName,
				secretNamespace,
				metric.name,
				metric.email,
				role,
			).Set(metric.durationUntilExpiry)
			JWTExpirationTimestamp.WithLabelValues(
				metric.algorithm,
				metric.audience,
				metric.subject,
				metric.id,
				scope,
				metric.issuer,
				keyName,
				secretName,
				secretNamespace,
				metric.name,
				metric.email,
				role,
			).Set(metric.expirationTimestamp)
			JWTIssuedAtTimestamp.WithLabelValues(
				metric.algorithm,
				metric.audience,
				metric.subject,
				metric.id,
				scope,
				metric.issuer,
				keyName,
				secretName,
				secretNamespace,
				metric.name,
				metric.email,
				role,
			).Set(metric.issuedAtTimestamp)
			JWTIssuedSinceSeconds.WithLabelValues(
				metric.algorithm,
				metric.audience,
				metric.subject,
				metric.id,
				scope,
				metric.issuer,
				keyName,
				secretName,
				secretNamespace,
				metric.name,
				metric.email,
				role,
			).Set(metric.durationSinceIssued)
		}
	}
	return nil
}

func (c *Exporter) ResetMetrics() {
	JWTExpirySeconds.Reset()
	JWTExpirationTimestamp.Reset()
	JWTIssuedAtTimestamp.Reset()
	JWTIssuedSinceSeconds.Reset()
}
