package schema

import (
	"database/sql"
	"flag"
	"fmt"
	"testing"

	. "github.com/pingcap/check"

	"github.com/guihouchang/go-mysql/client"
	_ "github.com/guihouchang/go-mysql/driver"
	"github.com/guihouchang/go-mysql/test_util"
)

var schema = flag.String("schema", "test", "MySQL Database")
var pwd = flag.String("pwd", "", "MySQL password")

func Test(t *testing.T) {
	TestingT(t)
}

type schemaTestSuite struct {
	conn  *client.Conn
	sqlDB *sql.DB
}

var _ = Suite(&schemaTestSuite{})

func (s *schemaTestSuite) SetUpSuite(c *C) {
	addr := fmt.Sprintf("%s:%s", *test_util.MysqlHost, *test_util.MysqlPort)

	var err error
	s.conn, err = client.Connect(addr, "root", *pwd, *schema)
	c.Assert(err, IsNil)

	s.sqlDB, err = sql.Open("mysql", fmt.Sprintf("root:%s@%s", *pwd, addr))
	c.Assert(err, IsNil)
}

func (s *schemaTestSuite) TearDownSuite(c *C) {
	if s.conn != nil {
		s.conn.Close()
	}

	if s.sqlDB != nil {
		s.sqlDB.Close()
	}
}

func (s *schemaTestSuite) TestSchema(c *C) {
	_, err := s.conn.Execute(`DROP TABLE IF EXISTS schema_test`)
	c.Assert(err, IsNil)

	str := `
        CREATE TABLE IF NOT EXISTS schema_test (
            id INT,
            id1 INT,
            id2 INT,
            name VARCHAR(256),
            status ENUM('appointing','serving','abnormal','stop','noaftermarket','finish','financial_audit'),
            se SET('a', 'b', 'c'),
            f FLOAT,
            d DECIMAL(2, 1),
            uint INT UNSIGNED,
            zfint INT ZEROFILL,
            name_ucs VARCHAR(256) CHARACTER SET ucs2,
            name_utf8 VARCHAR(256) CHARACTER SET utf8,
            name_char CHAR(10),
            name_binary BINARY(11),
            name_varbinary VARBINARY(12),
            PRIMARY KEY(id2, id),
            UNIQUE (id1),
            INDEX name_idx (name)
        ) ENGINE = INNODB;
    `

	_, err = s.conn.Execute(str)
	c.Assert(err, IsNil)

	ta, err := NewTable(s.conn, *schema, "schema_test")
	c.Assert(err, IsNil)

	c.Assert(ta.Columns, HasLen, 15)
	c.Assert(ta.Indexes, HasLen, 3)
	c.Assert(ta.PKColumns, DeepEquals, []int{2, 0})
	c.Assert(ta.IsPrimaryKey(0), IsTrue)
	c.Assert(ta.IsPrimaryKey(1), IsFalse)
	c.Assert(ta.IsPrimaryKey(2), IsTrue)
	c.Assert(ta.IsPrimaryKey(3), IsFalse)
	c.Assert(ta.GetPKColumn(0), Equals, &ta.Columns[2])
	c.Assert(ta.GetPKColumn(1), Equals, &ta.Columns[0])
	c.Assert(ta.GetPKColumn(2), IsNil)
	c.Assert(ta.GetPKColumn(3), IsNil)
	c.Assert(ta.Indexes[0].Columns, HasLen, 2)
	c.Assert(ta.Indexes[0].Name, Equals, "PRIMARY")
	c.Assert(ta.Indexes[2].Name, Equals, "name_idx")
	c.Assert(ta.Columns[3].Type, Equals, TYPE_STRING)
	c.Assert(ta.Columns[3].MaxSize, Equals, uint(256))
	c.Assert(ta.Columns[3].FixedSize, Equals, uint(0))
	c.Assert(ta.Columns[4].EnumValues, DeepEquals, []string{"appointing", "serving", "abnormal", "stop", "noaftermarket", "finish", "financial_audit"})
	c.Assert(ta.Columns[5].SetValues, DeepEquals, []string{"a", "b", "c"})
	c.Assert(ta.Columns[7].Type, Equals, TYPE_DECIMAL)
	c.Assert(ta.Columns[0].IsUnsigned, IsFalse)
	c.Assert(ta.Columns[8].IsUnsigned, IsTrue)
	c.Assert(ta.Columns[9].IsUnsigned, IsTrue)
	c.Assert(ta.Columns[10].Collation, Matches, "^ucs2.*")
	c.Assert(ta.Columns[10].MaxSize, Equals, uint(256))
	c.Assert(ta.Columns[10].FixedSize, Equals, uint(0))
	c.Assert(ta.Columns[11].Collation, Matches, "^utf8.*")
	c.Assert(ta.Columns[12].Type, Equals, TYPE_STRING)
	c.Assert(ta.Columns[12].MaxSize, Equals, uint(10))
	c.Assert(ta.Columns[12].FixedSize, Equals, uint(10))
	c.Assert(ta.Columns[13].Type, Equals, TYPE_BINARY)
	c.Assert(ta.Columns[13].MaxSize, Equals, uint(11))
	c.Assert(ta.Columns[13].FixedSize, Equals, uint(11))
	c.Assert(ta.Columns[14].Type, Equals, TYPE_BINARY)
	c.Assert(ta.Columns[14].MaxSize, Equals, uint(12))
	c.Assert(ta.Columns[14].FixedSize, Equals, uint(0))

	taSqlDb, err := NewTableFromSqlDB(s.sqlDB, *schema, "schema_test")
	c.Assert(err, IsNil)

	c.Assert(taSqlDb, DeepEquals, ta)
}

func (s *schemaTestSuite) TestQuoteSchema(c *C) {
	str := "CREATE TABLE IF NOT EXISTS `a-b_test` (`a.b` INT) ENGINE = INNODB"

	_, err := s.conn.Execute(str)
	c.Assert(err, IsNil)

	ta, err := NewTable(s.conn, *schema, "a-b_test")
	c.Assert(err, IsNil)

	c.Assert(ta.Columns[0].Name, Equals, "a.b")
}
