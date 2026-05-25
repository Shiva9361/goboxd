## [2026-05-23] [Context: Choosing a production Go web framework]
**Prompt:**
Best module suggestiond for webserver in go with multithreading and generally used in production
**Response summary:**
Explained that Go handles multithreading natively via goroutines, so no special module is needed for that. Recommended Chi (for standard library compatibility), Gin (for batteries-included), and Fiber (for raw performance) as production frameworks.
**What we used / didn't use:**
Ended up using gin for now


## [2026-05-25] [Context: Integrating nsjail sandboxing in Go]
**Prompt:**
Is there a nsjail api in GO directly if yes, provide me the documentation for it.
**Response summary:**
Clarified that there is no official native Go API for nsjail. Recommended using Go's standard `os/exec` package to shell out to the command-line binary, or checking out Chromium's internal `nsjailutil` wrapper.
**What we used / didn't use:**
Discarded the search for a native Go module and accepted the `os/exec` pattern to interface with the nsjail C++ binary directly.

## [2026-05-25] [Context: Go documentation conventions]
**Prompt:**
best way to write docstrings in go?
**Response summary:**
Detailed that Go uses standard `//` comments rather than Python-style docstrings. Outlined the golden rule: comments must start with the name of the function/struct and be written as complete sentences without `@param` tags.
**What we used / didn't use:**
Using the `//` format 

## [2026-05-25] [Context: Troubleshooting Viper config paths]
**Prompt:**
[Viper error] Config File "settings.yaml" Not Found in "[/config]"... config is in the root project dir.
**Response summary:**
Identified the difference between absolute paths (`/config`) and relative paths (`./config`) in Linux/Docker environments. And found that I had forgotten to copy the config folder into the docker env.
**What we used / didn't use:**
Used `./config` to accurately map to the project's config directory, discarding the absolute root path `/config` which was causing crashes and also added the config folder into docker's working env.

## [2026-05-25] [Context: Unmarshaling complex YAML into Go structs]
**Prompt:**
how to make a struct for a yaml file for unmarshall with viper
**Response summary:**
Provided the nested Go structs using `mapstructure` tags and explained how to unmarshal a YAML array/slice using `viper.UnmarshalKey`.
**What we used / didn't use:**
Used nested structs with `mapstructure` tags to handle snake_case YAML keys, discarding standard `yaml` struct tags.
