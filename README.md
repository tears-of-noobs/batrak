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


