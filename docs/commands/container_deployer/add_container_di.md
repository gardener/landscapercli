# Create and Add Container Deploy Item to a Component

This tutorial describes how to create a component with a blueprint and a container deploy item. With a container deploy 
item  you could provide your own docker image which is then executed by the landscaper. This way you could run whatever 
coding you need to set up your cloud environment.

It is helpful to read the [create tutorial](../create_component/create.md) first to better understand the basic 
concepts.

## Command to add a Container Deploy Item

You can add a container deploy item to a component with a blueprint with the following command:

```
landscaper-cli component add container deployitem \
  myDeployItem \
  --resource-version [version]
  --component-directory [some-path] \
  --image [oci-image-reference] \
  --import-param [param-name:param-type(optional, multi-value)] \
  --export-param [param-name:param-type (optional, multi-value)] \
  --command [command with modifier (optional, multi-value)] \
  --args [arguments (optional, multi-value)] \
  --cluster-param [target-cluster-param-name (optional)] 
```

The meaning of the arguments and flags is as follows:

- component-directory: Path to the component directory

- image: Reference to the OCI image which should be executed.

- import-param: The value of this flag consists of two parts separated by a colon, e.g. *replicas:integer*.
  The first part defines the name of an import parameter for the blueprint and the second part its type.

- export-param: The value of this flag consists of two parts separated by a colon, e.g. *replicas:integer*.
  The first part defines the name of an import parameter for the blueprint and the second part its type.

- command: The command to be executed in your image.

- args: Arguments for the command.

- cluster-param: Defines the name of the import parameter of the blueprint for access data to e.g. access 
  data to a target cluster.
  
We will describe the command and parameters in more detail in the following.

## An example Container Deploy Item

With a container deploy item you could run whatever you want in a container, e.g. a programm in go, 
C++ etc. Here we want to concentrate on the landscaper and container deployer concepts. Therefore, we want to run a 
very simple program which gets as input one word after the other and returns the sequence of words provided so far. 

### The Example Script (Program)

Next, we implement the program and build a docker image with it. The scripts looks as follows and could be also found 
in the file [scripts](resources/image-resources/script.sh):

```
#!/bin/sh

# read input parameter
IMPORTS=$(cat $IMPORTS_PATH)

# read sleep time before, only introduced to prolongate the command such that you could inspect the running
# container before the program is executed
sleepTimeBefore=$(echo $IMPORTS | jq .imports.sleepTimeBefore)
sleep $sleepTimeBefore

# if $OPERATION = "RECONCILE" append the input word, otherwise clean up
if [ $OPERATION == "RECONCILE" ]; then
   # read import parameter word and remove double quotes
  inputWord=$(echo $IMPORTS | jq .imports.word)
  inputWord=$(echo $inputWord | sed 's/"//g')

  # create empty state file if it does not exists
  [[ ! -f $STATE_PATH/state.txt ]] && echo "" > $STATE_PATH/state.txt

  # append word to state
  state=$(cat $STATE_PATH/state.txt)
  state="$state $inputWord"
  echo $state > $STATE_PATH/state.txt

  # write export values
  output="{\"sentence\": \"$state\"}"
  echo $output > $EXPORTS_PATH
else
  # do some cleanup, not really necessary because the state will not be kept anyhow
  rm $STATE_PATH/state.txt
fi

# read sleep time after, only introduced to prolongate the command such that you could inspect the running
# container before the program is executed
sleepTimeAfter=$(echo $IMPORTS | jq .imports.sleepTimeAfter)
sleep $sleepTimeAfter
```

When the script is executed by the landscaper in combination with the component deployer, the following environment 
variables are set:

- OPERATION: Contains either `RECONILE` or `DELETE` to inform the program if it should do its job or that the 
  deploy item was deleted, and a potential clean up should be executed.
  
- IMPORTS: Describes the path to a json file containing the input data for your program. You can find an example in
  the file [imports.json](./resources/misc/imports.json). In this file you find the following sections:
  
  - *blueprint*: contains the name of the blueprint
  - *cd*: the component descriptor
  - *componentDescriptorDef*: base information of the component
  - *components*: the resolved component descriptor list, which means that all transitive component descriptors are 
    included in a list
  - *imports*: a JSON structure containing the values of all import parameters. The input parameter are read 
  with `IMPORTS=$(cat $IMPORTS_PATH)`. In the example above we expect 4 input 
  parameters in the following format (of course more complex data is possible). In the example script, we do not use the
  target-cluster parameter, but we included it to illustrate how to access a target object. 
  
      ```yaml
      {
        "sleepTimeBefore": [some number],
        "sleepTimeAfter": [some number],
        "target-cluster": [target object including the kubeconfig]
        "word": [some word]
      }
      ```
  
- STATE_PATH: Contains the path to a folder, where you could store status information. Here you are not restricted to any
  format. You could create as many files with different data as you need. 
  - In our example we create here just one single file `state.txt` containing all word provided so far.
  
- EXPORTS_PATH: Contains the path to a file were the output data should be stored in json or yaml format. The
  file itself does not already exist. Only the parent folder is already there.
  - The example script just creates:
  ```
  {"sentence": word1 word2 ...}"
  ```

### Lifecycle

Now it is time to give you a rough understanding of the runtime lifecycle of programs specified as container deploy 
items. When an installation with our example container deploy item is created, the landscaper and container deployer 
acts as following:

- The program (the script in our example) is configured as a container deploy item in a blueprint and an installation 
  referencing this blueprint is deployed in a k8s cluster. The installation and teh blueprint define all 
  input parameter the program needs. We will show this in more detail later.

- If all input parameters of the installation are available, the container deployer starts the 
  program in a container in a `pod` providing the input data in the file specified by `$IMPORTS`. All 
  other environment variable described above are set accordingly. Especially `$OPERATION` is set on `RECONCILE`.
  
- When the program has finished the output data are fetched and provided as output data of the corresponding 
  installation. Furthermore, the data stored in the directory `STATE_PATH` are saved by the landscaper. Then, the
  `pod`, in which the container with our program was executed, is deleted.
  
- If new input data is provided to the installation (or something else has changed here), the program is started in a new
  pod with the latest input data available. Furthermore, the folder `STATE_PATH` contains exactly the data found here
  after the last run of the program. When the program has finished again, the new output data are fetched and provided
  by the installation, the state is saved and the `pod` deleted. This loops repeats until the installation with
  the container deploy item is deleted.
  
- If the installation is deleted, a container with our program is started as before. Only the variable `$OPERATION`
  now contains `DELETE` to inform the program that it is now time to clean up if necessary. Our example program
  does not install something. Therefore, cleanup is not relevant, we just remove the state file which is not really
  needed, because after a delete operation, the landscape will not keep these data anyhow.
  
### Create a Docker Image with our Example Script/Program

The landscaper or container deployer could not start our example script directly. It could just start it as a command 
of a docker image. Therefore, we need to create such a docker image for our script.

The docker file to build this image is in the file [Dockerfile](resources/image-resources/Dockerfile). It uses the 
alpine base image, installs `jq`, copies our script and sets its permissions.  

With the following commands executed on the [folder of the Dockerfile](resources/image-resources) you could build
and upload it to some OCI registry:

```shell
docker login ..."

docker build --tag your-registry/your-path/containerexample:0.1.0 .
docker push your-registry/your-path/containerexample:0.1.0
```

### Create a component

If you just want to copy and execute the example commands below with only small modifications, do the following steps:

- Clone the [Landscaper CLI Git repository](https://github.com/gardener/landscapercli).

- Define a variable *LS_ROOT_DIR* for the root directory of your cloned repository.

  ```
  export LS_ROOT_DIR=<path to the root directory of the landscapercli clone>
  ```

- Define a variable *LS_COMPONENT_DIR* for the directory in which you want to develop components. In this
  directory a subfolder for the example component will be created later.

  ```
  export LS_COMPONENT_DIR=<path to the directory of the demo component>
  ```

First off all we create a component skeleton with the following command:

```
landscaper-cli component create github.com/gardener/landscapercli/examplecontainer v0.1.0 \
  --component-directory $LS_COMPONENT_DIR/demo-component
```

An example of the component could be found in the folder [01-create-component](resources/01-create-component).

### An Example Container Deploy Item

Now we add the image with our script as a container deploy item to the component:

```
landscaper-cli component add container deployitem \
  examplecontainer \
  --resource-version 0.1.0 \
  --component-directory $LS_COMPONENT_DIR/demo-component \
  --image your-registry/your-path/containerexample:0.1.0 \
  --cluster-param target-cluster
  --import-param word:string \
  --import-param sleepTimeBefore:integer \
  --import-param sleepTimeAfter:integer \
  --export-param sentence:string \
  --command './script.sh'
```

The resulting component files could be found [here](resources/02-add-container-deploy-item/demo-component).
The command makes the following changes to the component:

- It adds two input parameter `sleepTimeBefore` and `sleepTimeBefore` of type integer and an input parameter 
  `word` of type string to the [blueprint.yaml](resources/02-add-container-deploy-item/demo-component/blueprint/blueprint.yaml). 
  Currently, only simple parameter of type string, integer or boolean are allowed for import parameter in the 
  landscaper CLI, but you could create more complex ones by manipulating the yaml files manually.
  It also adds a target import parameter `target-cluster`.
  
- It adds one export parameter `sentence` of type string to the 
  [blueprint.yaml](resources/02-add-container-deploy-item/demo-component/blueprint/blueprint.yaml).
  Currently, the landscaper CLI supports only export parameters of the elementary types string, integer or boolean. 
  However, you can create more complex ones by manipulating the yaml files manually.
  
- It adds an export execution to the blueprint. The export execution defines how the values of the export parameters are 
  retrieved from the output of the program executed in the container. The export execution generated by the landscaper CLI
  fetches the value of an export parameter from the corresponding property in the output of the program. 
  For example, in the current scenario there is an export parameter `sentence`, and the script executed in the container
  writes a JSON document `{"sentence": ...}` to the output file `$EXPORTS_PATH`. The value of the 
  export parameter `sentence` comes from the corresponding property `sentence` in the JSON output of the program.
  
- It creates the file [deploy-execution-examplecontainer.yaml](resources/02-add-container-deploy-item/demo-component/blueprint/deploy-execution-examplecontainer.yaml)
  with the definition of the new container deploy item. In this file you find:
  - A section `importValues` such that all import data will be available at runtime.
  - The reference to the OCI image.
  - The command (and optionally arguments) which will be executed in the container running the image. In our example 
    just the script `script.sh` is called without any arguments.
  
- It adds a reference to the file 
  [deploy-execution-examplecontainer.yaml](resources/02-add-container-deploy-item/demo-component/blueprint/deploy-execution-examplecontainer.yaml) 
  to the [blueprint.yaml](resources/02-add-container-deploy-item/demo-component/blueprint/blueprint.yaml)
  
- It adds the reference to the OCI image as a resource to the 
  [resources.yaml](resources/02-add-container-deploy-item/demo-component/resources.yaml).

### Upload the Component

As already described in [creates.md](../create_component/create.md#3-upload-component-into-an-oci-registry), we now 
add the base URL of the OCI registry and the resources to the component descriptor and upload the component to an 
OCI registry. 

### Create an Installation

Next we need an installation referencing the new component. An example could be found 
[here](resources/installations/installation.yaml). Be aware that you have to change the `baseUrl` to that of your
OCI registry. 

In the example we have set the input parameter `word` on *word1*, and the sleep times on 5 minutes. When you deploy
this installation on the landscaper cluster after a short time a pod is started executing the script. During that
time you could open a shell in the pod to analyze the settings with:

```
kubectl -n some-namespace exec --stdin --tty pod-name  -- /bin/sh
```

During the first 5 minutes you can already see the 
import data. After about five minutes you should also be able to see the output data in the file `$EXPORTS_PATH`.

## Todo

- Docu
  - integrate target-cluster
  - integrate other stuff like data in blueprint, component-descriptor, etc.

- Implementation
  - access to images in secured OCI registry
  
- integration tests