# Kawa

A tool for managing custom [Mura](https://murasoftware.com) modules.
- - -
## kawa-util

The util handles server side module management. Create manifests and serve the current manifests / serve the modules via an http file server.

Arguments:
```bash
scan    scan the current directory and create manifests.
-s      Start server
-d      Directory to serve
```
- - - 
## kawa (client)

The client handles most of the user interaction.

Arguments:

```bash
list                    List available modules
install <module name>   Download and unzip a module
remove <module name>    Removes a module
info <module name>      Displays information about a module
```
- - -
## mura-module.json

A Mura module requires a properly created mura-module.json file at the root of the module. You then zip the module to be served by the kawa-util server.

Components of a mura-module.json:

| Key | Description | Required |
| ----------- | ----------- | ----------- |
| name | name of the module | X |
| version | module version | X |
| description | brief description of the module | |
| author | author name or email or alias | |
| repo | link to the repository to direct a user to | |

### Example:

```json
{
    "name": "kawa",
    "version": "0.0.1",
    "description": "Kawa Module Example",
    "author": "Michael Hampton",
    "repo": "https://github.com/Michampt/kawa"
}
```

For a simple example module, see [Examples](https://github.com/Michampt/kawa/tree/main/kawa-utils/test/module-dirs)
