# awc.go &nbsp; [![](https://pkg.go.dev/badge/github.com/lus/awc.go/awc.svg)](https://pkg.go.dev/github.com/lus/awc.go/awc)

`awc.go` is a simple-to-use Go client for the Aviation Weather Center's [Text Data Server](https://aviationweather.gov/dataserver)
which provides, among other things, METAR and TAF data directly from a government's source.

## Usage

### Adding the dependency

First of all you have to add the dependency to your project:

```
go get -u github.com/lus/awc.go/awc
```

### Fetch METAR data

```go
// We need to build a query to use to fetch our METAR data.
// Please refer to the documentation on pkg.go.dev for more information.
query := new(awc.METARQuery).
	Station("EDDF").
	HoursBeforeNow(1).
	MostRecent(true)

// Then we need to execute the just-built query.
response, err := awc.GetMETAR(query)
if err != nil {
	panic(err)
}

// Please keep in mind that, only because the GetMETAR function did not return any error, the request may still be flawed. 
if len(response.Warnings) > 0 {
	fmt.Printf("API warning(s): %s\n", strings.Join(response.Warnings)))
}
if len(response.Errors) > 0 {
	panic(fmt.Sprintf("API error(s): %s", strings.Join(response.Errors)))
}

fmt.Println(response.METARs[0].RawText)
// -> "EDDF 262150Z VRB02KT CAVOK 02/M01 Q1033 NOSIG"
```

### Fetch TAF data

**TAF data fetching is not yet implemented.**
