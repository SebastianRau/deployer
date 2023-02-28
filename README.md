# deployer
Deploment Tool for git pulls and File Copy, Delete, etc. which uses json config files as recipe.
It also uilizes golang templates to empower the json config files.

## Usage
Define deployment steps, like git sync, copy, delete, write File, replace text in files in one json file and execute it on target system.

## Parameter
- -config.file: The json config file, which defines all  steps
- -config.data: The json data file, if golang templates are used
- -config.output: Settting an ouptut file will parse the config and file into the output file. The config will be been executed.
- -v: Verbose Output from each step

## Config file
The config file must be a json file which a array names steps.
Each step is an object with the following fileds.

- type:             string, mandatory. The Step Type, see [Step List](##Steps)
- parameter:        object, mandatory. Parameter for the defined step
- ignoreError:      boolean, optional. Stops execution of all step if an error occurs
- description:      string, optinal. Description will be shown during execution.

```
{
  "steps":[
    {
      "type":"gitUpdate",
      "description":"Updating BasicScripts Repo",
      "ignoreError":true,
      "parameter":{ }
    }
  ]
}
```

## Template data
a json object which contains all used varables for the template
[see golang template](https://pkg.go.dev/text/template)

### Template functions

``` add(a, b int) int ``` arithmetic addition 

``` sub(a, b int) int ```     arithmetic subtraction

``` hasNext(array []interface{}, idx int) bool ``` checks if index is smaller last array index

``` isLast(array []interface{}, idx int) bool  ``` check if index is equal to last arry index

Example:
Add , to all except the last entry 
```
{
"arr" :[
    {{range $i, $val := .array}}"some array {{ $val }}" {{if hasNext $.mapList $i}},{{end}}
    {{end}}
    ]
}
```


## Output file
just parses the given config and data file to the output file.

## Steps

    Name                        json Type
    CreateFolder                "mkdir"
    Copy                        "cp"
    Delete                      "rm"
    ReplaceText                 "replaceText"
    FileWriter                  "fileWriter"
    Exec                        "exec"
    EcecEachFile                "execEachFile"
    GitUpdate                   "gitUpdate"
    SubSteps                    "subSteps"
                      
### CreateFolder
Creates all given folder. 
``` 	
parameter: {
  "destination":[
    "./temp/foo/bar",
    "./temp/bla/bla/bla"
  ]
}
```



### Copy
Copy Files or Folders from source to destination.
[Wildcards](#wildcards) * and ** are allowed.
``` 	
parameter {
  "src_dest":{
    "source":"destination",
    "source2":"destination2"
  },
  "ignoreDotFiles":false
}
```

### Delete
Deletes Files or Folders.
only [Wildcards](#wildcards) * is allowed
``` 	
parameter: {
  "destination":[
    "./temp/foo/bar/*",
    "./temp/bla/bla/bla.go"
  ]
}
```
e.g Delete("./temp/") will delete the folder "temp" with all content

e.g Delete("./temp/*") will delete all the content, but not the folder itself

### ReplaceText
Replaces given placeholdes in a text File.
The file will be overwritten!
``` 	
parameter: {
    "destination": "foo.txt",
    "map":{
       "__MY_PLACEHOLDER1__" : "my new text1",
       "__MY_PLACEHOLDER2__" : "my new text2",
    }
  ]
}
```
### FileWriter
Creates or Overwrites a text File.
``` 	
parameter: {
    "filename": "foo.txt",
    "text":[
       "my new text line 1",
       "my new text line 2"
       ]
    }
  ]
}
```

### Exec                 
executes an external command
``` 	
parameter: {
    "command": "chmod",
    "path"; "/home/foo/bar"
    "params":[
       "+x",
       "runServer.sh"
       ]
    }
  ]
}
```



### EcecEachFile    
Executes an external command for file wich is found from the search parameter

Only [Wildcards](#wildcards) * in search is allowed

The following macros could be used in params array

```__FILE__``` will be replaced by filename e.g. "foo.pbo"

```__DIR__```will be replaced by the dir e.g "bla/bla/"

```__PATH__``` will be replaced by dir and filename e.g. "bla/bla/foo.pbo"

``` 	
parameter: {
    "command": "sign.exe",
    "search"; "/my/signed/gameMod/*.pbo"
    "path"; "/my/signed/gameMod/"

    "params":[
       "-sign foobar.keyfile",
       "__PATH__"
       ]
    }
  ]
}
```

### GitUpdate    
Clones or pulls a git repository via ssh from given url.


It uses a no auth (with https url) or a ssh private keyfile (use ssh url type!) for auth (e.g. a deployment key).

The option ```"errorOnNoUpdate": true``` will stop execution if ```"ignoreError"``` is set to ```false```.

```
parameter: {
    "url":              "git@github.com:SebastianRau/deployer.git",
    "dir":              "./deployer/",
    "keyfile":          "id_deploymentkey",
    "password":         "my secret password for keyfile"
    "errorOnNoUpdate":  false
    }
  ]
}
````


### SubSteps     
executes a nested json config file
```
parameter: {
    "config":            "another.json.tpl",
    "data":              "another.data.json",
    }
  ]
}
````


## Wildcards

### Wildcard *
The * wildcard gives all matching files in the folder, but not in subfolder.
```
temp/a.foo
temp/z.bar
temp/b.foo
temp/bla/c.foo
temp/bla/y.bar
temp/bla2/d.foo

```

``` "search": "temp/*.foo ```

results to:
```
temp/a.foo
temp/b.foo
```

### Wildcard **
The ** wildcard gives all matching files in the folder, AND in all subfolder.
```
temp/a.foo
temp/z.bar
temp/b.foo
temp/bla/c.foo
temp/bla/y.bar
temp/bla2/d.foo

```

``` "search": "temp/**.foo ```

results to:
```
temp/a.foo
temp/b.foo
temp/bla/c.foo
temp/bla2/d.foo
```

