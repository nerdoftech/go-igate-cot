package igate

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
)

var (
	testPyld = []byte("KI4ABC>AB1CDE,WIDE2*,qAO,KI1ABC:!3223.23N/09988.99W-")
)

func TestIgate(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Igate Suite")
}
