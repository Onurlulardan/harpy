package main

import (
	"fmt"
	"net/url"
	"time"
)

func addConnectionArtifacts(set *artifactSet, _ time.Time) error {
	passwords, err := newConnectionPasswords()
	if err != nil {
		return fmt.Errorf("generate connection passwords: %w", err)
	}
	accessKey, err := randomAccessKeyID()
	if err != nil {
		return err
	}
	secretKey, err := randomBase64(32)
	if err != nil {
		return err
	}
	rootAccessKey, rootSecretKey, err := newMinIORootCredentials(accessKey, secretKey)
	if err != nil {
		return err
	}
	rabbitMQConfig, err := buildRabbitMQConfig(passwords.rabbitmq)
	if err != nil {
		return err
	}
	files := map[string]string{
		"postgres-bootstrap-password.txt": passwords.bootstrap,
		"postgres-bootstrap-dsn.txt": postgresDSN(
			bootstrapPrincipal(set), passwords.bootstrap, "postgres"),
		"postgres-runtime-password.txt":         passwords.runtime,
		"postgres-migrator-password.txt":        passwords.migrator,
		"postgres-cluster-admin-password.txt":   passwords.clusterAdmin,
		"postgres-extension-admin-password.txt": passwords.extensionAdmin,
		"postgres-runtime-dsn.txt": postgresDSN(
			"harpy_v2_runtime", passwords.runtime, "harpy"),
		"postgres-migrator-dsn.txt": postgresDSN(
			"harpy_migrator", passwords.migrator, "harpy"),
		"postgres-reactivation-observer-password.txt": passwords.observer,
		"postgres-cluster-admin-maintenance-dsn.txt": postgresDSN(
			"harpy_cluster_admin", passwords.clusterAdmin, "postgres"),
		"postgres-cluster-admin-target-dsn.txt": postgresDSN(
			"harpy_cluster_admin", passwords.clusterAdmin, "harpy"),
		"postgres-extension-admin-dsn.txt": postgresDSN(
			"harpy_extension_admin", passwords.extensionAdmin, "harpy"),
		"rabbitmq-password.txt":     passwords.rabbitmq,
		"rabbitmq-url.txt":          rabbitMQURL("harpy_worker", passwords.rabbitmq),
		"rabbitmq.conf":             rabbitMQConfig,
		"s3-access-key-id.txt":      accessKey,
		"s3-secret-access-key.txt":  secretKey,
		"minio-root-access-key.txt": rootAccessKey,
		"minio-root-secret-key.txt": rootSecretKey,
		"minio.env": "MINIO_ROOT_USER=" + rootAccessKey + "\n" +
			"MINIO_ROOT_PASSWORD=" + rootSecretKey + "\n",
		"smtp-password.txt": passwords.smtp,
		"smtp-auth.txt":     "harpy-local:" + passwords.smtp + "\n",
	}
	for name, value := range files {
		if err := set.add(name, []byte(value)); err != nil {
			return err
		}
	}
	return addInvitationTemplate(set)
}

func bootstrapPrincipal(set *artifactSet) string {
	if set.options.profile == profileForgeGuard {
		return "harpy_bootstrap"
	}
	return "local_bootstrap_admin"
}

func newMinIORootCredentials(
	runtimeAccessKey string,
	runtimeSecretKey string,
) (string, string, error) {
	for {
		accessKey, err := randomAccessKeyID()
		if err != nil {
			return "", "", err
		}
		secretKey, err := randomBase64(32)
		if err != nil {
			return "", "", err
		}
		if accessKey != runtimeAccessKey && secretKey != runtimeSecretKey {
			return accessKey, secretKey, nil
		}
	}
}

type connectionPasswords struct {
	bootstrap, runtime, migrator string
	clusterAdmin, extensionAdmin string
	observer, rabbitmq, smtp     string
}

func newConnectionPasswords() (connectionPasswords, error) {
	values := make([]string, 8)
	for index := range values {
		value, err := randomBase64(32)
		if err != nil {
			return connectionPasswords{}, err
		}
		values[index] = value
	}
	return connectionPasswords{
		bootstrap: values[0], runtime: values[1], migrator: values[2],
		clusterAdmin: values[3], extensionAdmin: values[4],
		observer: values[5], rabbitmq: values[6], smtp: values[7],
	}, nil
}

func postgresDSN(username, password, database string) string {
	value := url.URL{Scheme: "postgres", User: url.UserPassword(username, password),
		Host: "postgres:5432", Path: "/" + database}
	query := url.Values{}
	query.Set("sslmode", "disable")
	value.RawQuery = query.Encode()
	return value.String()
}

func rabbitMQURL(username, password string) string {
	return (&url.URL{Scheme: "amqp", User: url.UserPassword(username, password),
		Host: "rabbitmq:5672", Path: "/harpy"}).String()
}

func addInvitationTemplate(set *artifactSet) error {
	return addJSON(set, "invitation-template.json", map[string]any{
		"version": 1,
		"subject": "You're invited to ForgeCRUD",
		"text":    "Join your ForgeCRUD workspace: {{INVITATION_URL}}",
		"html":    invitationGraphiteHTML,
	})
}
