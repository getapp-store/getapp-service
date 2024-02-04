package bidding

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"ru/kovardin/getapp/app/modules/mediation/models"
	"ru/kovardin/getapp/app/modules/mediation/networks"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type BanditTestSuite struct {
	suite.Suite
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *BanditTestSuite) SetupTest() {
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *BanditTestSuite) TestBid() {
	bidding := Bidding{
		networks: cpmsMock{
			cpms: []models.CpmByNetwork{
				{
					1, 0.38,
				},
				{
					2, 0, // yandex
				},
				{
					4, 1.738,
				},
			},
		},
	}

	// нужно так как в коде используется рандом
	rand.Seed(1)

	var (
		val float64
		err error
	)

	// подобрано вручную для прохождения теста
	for i := 0; i < 9; i++ {
		val, err = bidding.Bandit(0.01, models.Unit{
			Network: models.Network{
				Name: networks.Yandex,
			},
		})
	}

	suite.T().Logf("value: %f, err: %+v", val, err)

	assert.Equal(suite.T(), 100.01, val)
	assert.NoError(suite.T(), err)

	//suite.Equal(5, suite.VariableThatShouldStartAtFive)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestBanditTestSuite(t *testing.T) {
	suite.Run(t, &BanditTestSuite{})
}

type cpmsMock struct {
	cpms []models.CpmByNetwork
}

func (m cpmsMock) CpmsByNetwork(from, to time.Time) ([]models.CpmByNetwork, error) {
	return m.cpms, nil
}
