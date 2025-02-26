## Run-manual

### Running the HAC Node

#### First, build and start the Hetu Chaos Chain node:

```
cd   hac-node
make build
```

#### Then, start the hac-node:
```
cd   script/
chmod +x local-hac-nodes.sh
./local-hac-nodes.sh
```

### Running the Agent Client

#### In a new terminal, build and run the sample applications:
```
cd samples/
make build
cd ./build
./samples start         
```
This will start the agent client. You can monitor the output to see the system's operation.


#### Creating a New Proposal in samples/build

To create a new proposal using the sample client:
```
./samples   propose --title "First agent title" --data "hellow workshop"   
```
This command sends a new proposal to the chain with the specified title and data.