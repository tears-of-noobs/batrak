# batrak
Jira issues worklog console helper

# Usage

First create configuration file in you home directory

```
vim ~/.batrakrc
```

and fill it with the following parameters

```
username = "JIRA_USERNAME"
password = "JIRA_PASSWORD"
jira_api_url = "http://JIRA.DOMAIN/rest/api/2"
project_name = "JIRA_PROJECT_NAME"
```
also, if you know filter ID in JIRA you may define it in config.
```
filter_id = JIRA_FILTER_ID
```

You may manually ordering print issues by status.
Just add following lines in you config with status and order description
```
[workflow]
  [[workflow.stage]]
    name = "In progress"
    order = 1
  [[workflow.stage]]
    name = "Analysis"
    order = 2
    .
    .
    .
    .
```
Batrak support hooks (pre_start, post_start, pre_stop, post_stop)
Hook - it just binary file or script that takes two string arguments:
* Jira issue Key - "TEST-100"
* Credential for connecting to JIRA API - "JIRA_USERNAME\*JIRA_PASSWORD\*JIRA_API_URL"

If you write you own hook, who does something with your issue, and you want use it after issue was stopped, 
add this lines in your config

```
[hooks]
post_stop = ["hook_name"]
```


### Commands

#### List 10 last issues assigne to JIRA_USERNAME

```
batrak -L
```

#### Show issue (Name, Status, Description)


```
batrak -Ln TEST-100
```

or if you watch issue in your JIRA_PROJECT_NAME

```
batrak -Ln 100
```

#### Show comments 

```
batrak -LCn TEST-100
```

#### Show available transitions for issue 

```
batrak -Mn TEST-100
```

#### Move issue 

```
batrak -Mn TEST-100 TRANSITION_ID
``` 

#### Start issue 

```
batrak -Sn TEST-100
``` 


#### Stop issue with logging work

```
batrak -Tn TEST-100
``` 


