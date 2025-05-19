package testparamssend

type EmbedStruct struct {
	NumberTest  float64 `json:"numbertest"  jsonschema:"numbertest" title:"Number Test" default:"1" require:"true" description:"test of number"`
	IntegerTest int     `json:"integerTest" jsonschema:"integerTest" title:"Integer Test" default:"1" require:"true" description:"test of Integer"`
	StringTest  string  `json:"stringTest"  jsonschema:"stringTest" title:"String Test" default:"1" require:"true" description:"test of string"`

	StringArrTest []string `json:"stringArrTest"  jsonschema:"stringArrTest" title:"String Array Test" default:"1" require:"true" description:"test of string array"`
}

type Config struct {
	NumberTest  float64 `json:"numbertest"  jsonschema:"numbertest" title:"Number Test" default:"1" require:"true" description:"test of number"`
	IntegerTest int     `json:"integerTest" jsonschema:"integerTest" title:"Integer Test" default:"1" require:"true" description:"test of Integer"`
	StringTest  string  `json:"stringTest"  jsonschema:"stringTest" title:"String Test" default:"1" require:"true" description:"test of string"`

	StringArrTest []string `json:"stringArrTest"  jsonschema:"stringArrTest" title:"String Array Test" default:"1" require:"true" description:"test of string array"`

	EmbedStruct      EmbedStruct   `json:"embedStructTest"  jsonschema:"embedStructTest" title:"Embed Struct Test" require:"true" description:"test of embed struct"`
	EmbedArrayStruct []EmbedStruct `json:"embedArrayStruct"  jsonschema:"embedArrayStruct" title:"Embed Array Struct Test" require:"true" description:"test of embed struct array"`
}

type TestDeployReturn struct {
	Config
}
