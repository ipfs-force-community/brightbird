package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func TestListDatabase(t *testing.T) {
	t.Skipf("local test ")
	format := "root:Aa123456@(192.168.200.175:3306)/%s?parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci&readTimeout=10s&writeTimeout=10s"
	databasePrix := "aaaabbbbccccc-"

	rand.Seed(time.Now().Unix())
	for i := 0; i < 10; i++ {
		err := CreateDatabase(fmt.Sprintf(format, databasePrix+strconv.Itoa(rand.Int())))
		require.NoError(t, err)
	}

	databases, err := ListDatabase(format)
	require.NoError(t, err)
	for _, db := range databases {
		if strings.Contains(db, databasePrix) {
			err = DropDatabase(fmt.Sprintf(format, db))
			require.NoError(t, err)
		}
	}
}
