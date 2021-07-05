$result = [Newtonsoft.Json.Linq.JObject]::new()
foreach($i in (gci -recurse -file './pkg/runner/testdata/event-types/')){
    $data = [Newtonsoft.Json.Linq.JObject]::Parse((get-content $i.FullName -raw))
    $result.Merge($data)
}
$Result.ToString() > merged.json
