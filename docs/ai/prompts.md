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

## [2026-05-28] [Context: Writing Tests in Go]
**Prompt:**
how to write tests in go?
**Response summary:**
Explained that Go uses the `testing` package for writing tests. Test functions must start with `Test` and take a `*testing.T` parameter. Provided an example of a simple test function that checks if a function returns the expected result.
**What we used / didn't use:**
Used the `testing` package and followed the convention of naming test functions with a `Test` prefix. Discarded any third-party testing frameworks for now, sticking to Go's built-in capabilities.

## [2026-05-28] [Context: Generating json files for testing]
**Prompt:**
Generate 5 json files for languages python, c, c++ for using in the testing 
**Response summary:**
Generated the required JSON files with varied code for each language testing
**What we used / didn't use:**
Given that the test files were exactly what i needed, I used them directly. Would have taken a lot more time for me to write them by hand.

## [2026-05-30] [Context: Generating curl request for testing flag-allowlist]
**Prompt**
Generate a curl request for c++ for the server api that we already discussed about to test it's flag allowlist by adding potentially dangerous flags
**Response summary:**
Generated the required request
**What we used / didn't use:**
Used the request as is and also added it as a test case in cpp 