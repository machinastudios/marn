# Marn - Yarn for Maven

Marn is a CLI tool that reads scripts from `pom.xml` and executes them, similar to how `yarn` works with `package.json`.

## Installation

### Global Installation

```bash
cd /path/to/marn
./marn init
```

This will create a symlink in `/usr/local/bin/marn` (standard Ubuntu PATH location) pointing to the currently running script. You may be prompted for `sudo` permissions.

After installation, you can use `marn` from anywhere in your system!

## Usage

After installation, you can use `marn` with the following commands:

- `marn init` - Install marn globally (creates symlink in /usr/local/bin)
- `marn install` - Install dependencies (mvn dependency:resolve)
- `marn link` - Link current project to local Maven repository (~/.m2)
- `marn install-deps` - Install dependencies (mvn dependency:resolve)
- `marn build` - Build the project (mvn clean compile)
- `marn test` - Run tests (mvn test)
- `marn package` - Package the project (mvn package)
- `marn run` - Build and run the JAR
- `marn clean` - Clean the project (mvn clean)
- `marn watch` - Watch for changes and rebuild
- `marn <script>` - Run custom script from pom.xml

### Linking Projects

If you're working on a local dependency (like `mshared`), use `marn link` to install it to your local Maven repository:

```bash
cd mshared
marn link
```

This runs `mvn clean install -DskipTests`, making the project available to other projects that depend on it.

## Custom Scripts

Define custom scripts in your `pom.xml` under `<properties>`:

```xml
<properties>
    <script.dev>mvn clean compile -DskipTests</script.dev>
    <script.lint>mvn checkstyle:check</script.lint>
</properties>
```

Then run them with:

```bash
marn dev
marn lint
```

## Watch Mode

Configure watch mode in your `pom.xml`:

```xml
<properties>
    <watch.dirs>src/main/java src/main/resources</watch.dirs>
    <watch.buildCommand>mvn compile</watch.buildCommand>
    <watch.skipTests>true</watch.skipTests>
    <watch.debounceTime>2</watch.debounceTime>
    <watch.postCommand>./marn run</watch.postCommand>
    <watch.localDeps>../mshared</watch.localDeps>
</properties>
```

Then run:

```bash
marn watch
```

This will watch for changes in the specified directories and rebuild automatically.
