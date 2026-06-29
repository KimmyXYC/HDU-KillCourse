package config

import "testing"

func TestApplyDefaults_WebDefaults(t *testing.T) {
	cfg := &Config{}
	applyDefaults(cfg)

	if cfg.Web.Host != "127.0.0.1" {
		t.Fatalf("expected default web host %q, got %q", "127.0.0.1", cfg.Web.Host)
	}
	if cfg.Web.Port != 6688 {
		t.Fatalf("expected default web port %d, got %d", 6688, cfg.Web.Port)
	}
}

func TestApplyDefaults_WebKeepsProvidedValues(t *testing.T) {
	cfg := &Config{Web: WebConfig{Host: "0.0.0.0", Port: 7788}}
	applyDefaults(cfg)

	if cfg.Web.Host != "0.0.0.0" {
		t.Fatalf("expected web host to be preserved, got %q", cfg.Web.Host)
	}
	if cfg.Web.Port != 7788 {
		t.Fatalf("expected web port to be preserved, got %d", cfg.Web.Port)
	}
}

func TestApplyDefaults_WebAuthDefaults(t *testing.T) {
	cfg := &Config{}
	applyDefaults(cfg)

	if cfg.Web.Auth.Enabled != "1" {
		t.Fatalf("expected default web auth enabled %q, got %q", "1", cfg.Web.Auth.Enabled)
	}
	if cfg.Web.Auth.Username != "admin" {
		t.Fatalf("expected default web auth username %q, got %q", "admin", cfg.Web.Auth.Username)
	}
	if cfg.Web.Auth.Password != "password" {
		t.Fatalf("expected default web auth password %q, got %q", "password", cfg.Web.Auth.Password)
	}
}

func TestApplyDefaults_WebAuthAllowsExplicitEmptyCredentials(t *testing.T) {
	cfg := &Config{Web: WebConfig{Auth: WebAuthConfig{Enabled: "1", Username: "", Password: ""}}}
	applyDefaults(cfg)

	if cfg.Web.Auth.Enabled != "1" {
		t.Fatalf("expected web auth enabled to be preserved, got %q", cfg.Web.Auth.Enabled)
	}
	if cfg.Web.Auth.Username != "" {
		t.Fatalf("expected web auth username to stay empty, got %q", cfg.Web.Auth.Username)
	}
	if cfg.Web.Auth.Password != "" {
		t.Fatalf("expected web auth password to stay empty, got %q", cfg.Web.Auth.Password)
	}
}
