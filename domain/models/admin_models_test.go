package models

import (
	"reflect"
	"testing"
)

func TestAdminUserModel_HasStatusField(t *testing.T) {
	userType := reflect.TypeOf(User{})
	field, ok := userType.FieldByName("Status")
	if !ok {
		t.Fatalf("expected User model to have Status field")
	}
	if got := field.Type.Kind(); got != reflect.String {
		t.Fatalf("expected User.Status to be string kind, got %s", got)
	}
}

func TestAdminAuditLog_TableName(t *testing.T) {
	var log AdminAuditLog
	if got := log.TableName(); got != "admin_audit_logs" {
		t.Fatalf("expected table name admin_audit_logs, got %s", got)
	}
}
