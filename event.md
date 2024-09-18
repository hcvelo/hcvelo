---
title: {{ .Title }}
weight: {{ .Weight }}
params:
    date: {{ .EventDate }}
    time: {{ .EventTime }}
    event_date: {{ .EventDate }}
    event_time: {{ .EventTime }}
    strava_url: {{ .URL }}
---

## {{ .Title }} 

{{ .Description }}

### Date

{{ .EventDate }} : {{ .EventTime }}

### Start

{{ .Address }}


