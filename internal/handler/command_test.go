package handler

import (
	"os"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestParseWriteParamsWrongParamName(t *testing.T) {
	profileName := "notFound"
	roMap := roSliceToMap([]models.ResourceOperation{{Index: ""}})
	params := "{ \"key\": \"value\" }"

	_, err := parseWriteParams(profileName, roMap, params)

	if  err == nil {
		t.Error("expected error")
	}
}

func TestParseWriteParamsNoParams(t *testing.T) {
	profileName := "notFound"
	roMap := roSliceToMap([]models.ResourceOperation{{Index: ""}})
	params := "{ }"

	_, err := parseWriteParams(profileName, roMap, params)

	if  err == nil {
		t.Error("expected error")
	}
}
