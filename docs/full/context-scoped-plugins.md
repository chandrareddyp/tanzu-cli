# Context-Scope Plugin Installation

## Abstract

The Tanzu CLI is an amalgamation of all the Tanzu infrastructure elements under
one unified core CLI experience. The core CLI supports a plugin model where the
developers of different Tanzu services (bundled or SaaS) can distribute plugins
that target the functionalities of the services they own. When users switch between
different services via the CLI context, we want to surface only the relevant
plugins for the given context for a crisp user experience.

When a user is working with multiple instances of a product, we want to
automatically select the right set of plugins and plugin versions for use with
based on the active context user is connected to.

The goals of Context-Scoped Plugin Installation are:

As a plugin developer, I want to,

- Recommend relevant plugins and their versions that might be needed by the user once the user creates or login to a context.

As a user, I want to,

- Install the recommended version of all plugins for the current context during context create
- Install the recommended version of any missing plugin for the current context via a sync command
- Upgrade all installed plugins to newer versions via a sync command, if the installed versions are not supported anymore
- Avoid re-downloading a plugin version if already present (e.g., if the same version of the package plugin is provided by two management clusters, do not re-download it)

## Plugin Discovery and Distribution

Discovery is the interface to fetch the list of available plugins, their
supported versions, and how to download them. For MVP, Tanzu CLI will have
a default plugin discovery source configured by default which returns a list
of available plugins. In the future, this can be made configurable.

Distribution is the interface to download a plugin binary for a given OS
and architecture combination. The discovery source will provide details about
the distribution regarding where to fetch the plugin binary.

Plugin availability is solely dependent on the configured discovery sources in the
tanzu configuration file. Each discovery source points to a plugin repository
which can contain one or multiple plugins.

## Standalone Plugins

The scope of a plugin depends on how the plugin is getting installed on the user's machine.
Users can run the `tanzu plugin search` command to see all available plugins from
the configured discovery sources.

If the user wants to install a plugin that is not dependent on any active context and
wants to use it with the Tanzu CLI, the user can run `tanzu plugin install <plugin-name>`
command to install the required plugin. Installing the plugin this way will make the
plugin a standalone plugin and it will not be associated with any contexts.

## Context-scoped Plugins

As mentioned above in the abstracts section, there might be a scenario when a user
is working with multiple contexts at a time and wants to automatically select the
right set of plugins and plugin versions based on the active context user is connected to.
The context-scoped plugin implementation is useful in this scenario.

When the CLI user creates a new context for the Tanzu CLI using the
`tanzu context create` command, CLI adds a context in the tanzu configuration file
and marks the newly created context as an active context for the specified target.

Now, this newly created context can also recommend the list of plugins and their versions
that are needed to be installed on the user's machine to interact with the created context.
Tanzu CLI automatically detects the list of recommended plugins and their versions and
installs them as part of the `tanzu context create` or `tanzu context use` command. Below is the workflow of
context-scoped plugin installation:

- User runs the `tanzu context create` or `tanzu context use` command to create a new context or switch active context
- Tanzu CLI gets the list of recommended plugins and their version from the created context
- Tanzu CLI finds the plugins and the plugin metadata in the available plugin list generated from the configured discovery sources
- Tanzu CLI fetches the plugin binary for these plugins from the specified location and installs the plugin

Users should understand that these plugins (that are installed based on the contexts) are
only available when the said context is active. If a user deletes the context the plugins
installed based on the previous context are no longer available to use with the CLI.
If the user switches the context to a different context using the `tanzu context use` command,
CLI will automatically install/update the recommended plugins based on the new context.

## Plugin Recommendations from Context

This section provides more details on how a context can provide more details about
recommended plugins to use when a user creates or activates a context.

### When the context is of type Kubernetes

When the context is of type kubernetes, Tanzu CLI uses kubernetes discovery to fetch the
list of recommended plugins and their versions. Using the kubernetes discovery implementation
Tanzu CLI queries the `CLIPlugin` resources available on the kubernetes cluster.

For example, If the user is expected to use the plugins `cluster:v1.0.0` and `feature:v1.2.0`
when talking to kubernetes cluster `test-cluster` then the cluster should have below
`CLIPlugin` resources defined:

```yaml
apiVersion: cli.tanzu.vmware.com/v1alpha1
kind: CLIPlugin
metadata:
  name: cluster
spec:
  recommendedVersion: v1.0.0
  description: Kubernetes cluster operations
```

```yaml
apiVersion: cli.tanzu.vmware.com/v1alpha1
kind: CLIPlugin
metadata:
  name: feature
spec:
  recommendedVersion: v1.2.0
  description: Feature plugin operations
```

### When the context is of type Mission-Control

When the context is of type mission control, Tanzu CLI uses REST discovery to fetch the
list of recommended plugins and their versions. Using the REST discovery implementation
Tanzu CLI queries the `<server-url>/v1alpha1/system/binaries/plugins` REST API that
should return a list of `CLIPlugin` information.