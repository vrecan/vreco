# Overview
We needed to start building out python data science projects dealing with complex dependencies like Pandas, Scikit-learn & Tensorflow. While trying to build these dependencies I have found that it can be pretty difficult just trying to use pip install with some of them. We had read that Anaconda / Mini-Conda / mamba / micromamba can make things a lot easier. We spent a lot of time bouncing between solutions and I wanted to document where we ended up as a team.

# Repo setup / Continuous integration
We use a mono repo with a projects folder that has every service split into their own folder.

```
/<company>/projects/<project-name>/
```

This allows us to leverage github CI workflows that only run checks based on file changes in paths.

Example github workflow file:

```yaml
name: <project-name> Checks
on:
  push:
    branches: [main]
    paths:
      - "projects/<project-name>/**/*"
  pull_request:
    branches: [main]
    paths:
      - "projects/<project-name>/**/*"
defaults:
  run:
    shell: bash
    working-directory: projects/<project-name>

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  type_and_lint_checking:
    name: check (${{ matrix.python-version }}, ${{ matrix.os }})
    runs-on: ${{ matrix.os }}
    timeout-minutes: 30
    strategy:
      fail-fast: false
      matrix:
        os: ["ubuntu-latest", "macos-latest"]
        python-version: ["3.9"]
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-python@v2
        with:
          python-version: 3.9
      - uses: mamba-org/provision-with-micromamba@main
        with:
          channel-priority: disabled
          environment-file: projects/<project-name>/environment.yml
      - name: Bash
        shell: bash -l {0}
        run: pyright .
      - name: Run PyLint
        shell: bash -l {0}
        run: pylint main.py pkg/
      - name: Run PyTest
        shell: bash -l {0}
        run: pytest .
  format_checking:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-python@v2
        with:
          python-version: 3.9
      - uses: psf/black@stable
        with:
          version: "22.3.0"

```
## Dependency issues
In the yaml file you will notice that we run checks on macos and ubuntu. Our production system runs on ubuntu but a lot of devs like to run things on their mac laptops. We found that deps can break in unexpected ways depending on the platform. By running things in OSX as well in CI we can find a lot of these issues before we merge.

Example environment.yml file:
```yml
name: <project-name>
channels:
  - conda-forge
  - anaconda
  - defaults
dependencies:
  - pandas=1.4.2
  - geopandas=0.11.0
  - numpy=1.21.5
  - pylint=2.12.2
  - black=22.3.0
  - python=3.9
  - types-mypy-extensions=0.4.19
  - pandas-stubs=1.2
  - pytest=7.1.1
  - fsspec=2022.5.0
  - gcsfs=2022.5.0
  - fastapi=0.78.0
  - uvicorn=0.17.6
  - pyarrow=8.0.0
  - google-cloud-storage=2.4.0
  - google-cloud-bigquery=3.2.0
  - google-cloud-logging=3.3.1
  - pandas-gbq=0.17.6
  - pyright=1.1.255
  - great-expectations==0.15.11
  - postal=1.1.9
  - pydantic=1.9.1
  - python-dotenv=0.20.0
  - geojson=2.5.0
  - swifter=1.3.4
  - pip=22.3.1
  - pip:
      - pyairtable==1.4.0
      - googlemaps==4.8.0
```

You can see that in this situation we actually use PIP inside of conda to pull in some deps that don't exist within the anacodna ecosystem.

Here is an example cloudbuild.yaml that works with this configuration:
```yaml
steps:
  # Get previous image
  - name: "gcr.io/cloud-builders/docker"
    entrypoint: "bash"
    args:
      [
        "-c",
        "docker pull us-central1-docker.pkg.dev/${PROJECT_ID}/<GCLOUD-PROJECT-NAME>/<PROJECT-NAME>:build || exit 0",
      ]
  - name: "gcr.io/cloud-builders/docker"
    # builds the environment file / mamba step of the build
    args:
      [
        "build",
        "-f",
        "Dockerfile",
        "--cache-from",
        "us-central1-docker.pkg.dev/${PROJECT_ID}/<GCLOUD-PROJECT-NAME/<PROJECT-NAME>:build",
        "--target",
        "build",
        "-t",
        "us-central1-docker.pkg.dev/${PROJECT_ID}/<GCLOUD-PROJECT-NAME/<PROJECT-NAME>:build",
        ".",
      ]

  # Docker Build
  - name: "gcr.io/cloud-builders/docker"
    args:
      [
        "build",
        "-t",
        "us-central1-docker.pkg.dev/${PROJECT_ID}/<GCLOUD-PROJECT-NAME/<PROJECT-NAME>:latest",
        "--cache-from",
        "us-central1-docker.pkg.dev/${PROJECT_ID}/<GCLOUD-PROJECT-NAME/<PROJECT-NAME>:build",
        ".",
      ]
images:
  [
    "us-central1-docker.pkg.dev/${PROJECT_ID}/<GCLOUD-PROJECT-NAME/<PROJECT-NAME>:latest",
    "us-central1-docker.pkg.dev/${PROJECT_ID}/<GCLOUD-PROJECT-NAME/<PROJECT-NAME>:build",
  ]
timeout: 1800s
options:
  machineType: "E2_HIGHCPU_8"

```


## Testing & Linting
We leverage pylint, pyright and pytest to test and validate all of our code in every project. 

Here is our pylint config:
```
[MASTER]

# module doctring and function docstring
disable=C0114,C0115,C0116,E1136,R0903

# A comma-separated list of package or module names from where C extensions may
# be loaded. Extensions are loading into the active Python interpreter and may
# run arbitrary code.
extension-pkg-allow-list=

# A comma-separated list of package or module names from where C extensions may
# be loaded. Extensions are loading into the active Python interpreter and may
# run arbitrary code. (This is an alternative name to extension-pkg-allow-list
# for backward compatibility.)
extension-pkg-whitelist=

# Return non-zero exit code if any of these messages/categories are detected,
# even if score is above --fail-under value. Syntax same as enable. Messages
# specified are enabled, while categories only check already-enabled messages.
fail-on=

# Specify a score threshold to be exceeded before program exits with error.
fail-under=10.0

# Files or directories to be skipped. They should be base names, not paths.
ignore=CVS

# Add files or directories matching the regex patterns to the ignore-list. The
# regex matches against paths and can be in Posix or Windows format.
ignore-paths=

# Files or directories matching the regex patterns are skipped. The regex
# matches against base names, not paths. The default value ignores emacs file
# locks
ignore-patterns=^\.#

# Python code to execute, usually for sys.path manipulation such as
# pygtk.require().
#init-hook=

# Use multiple processes to speed up Pylint. Specifying 0 will auto-detect the
# number of processors available to use.
jobs=1

# Control the amount of potential inferred values when inferring a single
# object. This can help the performance when dealing with large functions or
# complex, nested conditions.
limit-inference-results=100

# List of plugins (as comma separated values of python module names) to load,
# usually to register additional checkers.
load-plugins=

# Pickle collected data for later comparisons.
persistent=yes

# Minimum Python version to use for version dependent checks. Will default to
# the version used to run pylint.
py-version=3.8

# Discover python modules and packages in the file system subtree.
recursive=no

# When enabled, pylint would attempt to guess common misconfiguration and emit
# user-friendly hints instead of false-positive error messages.
suggestion-mode=yes

# Allow loading of arbitrary C extensions. Extensions are imported into the
# active Python interpreter and may run arbitrary code.
unsafe-load-any-extension=no


[MESSAGES CONTROL]

# Only show warnings with the listed confidence levels. Leave empty to show
# all. Valid levels: HIGH, CONTROL_FLOW, INFERENCE, INFERENCE_FAILURE,
# UNDEFINED.
confidence=

# Disable the message, report, category or checker with the given id(s). You
# can either give multiple identifiers separated by comma (,) or put this
# option multiple times (only on the command line, not in the configuration
# file where it should appear only once). You can also use "--disable=all" to
# disable everything first and then re-enable specific checks. For example, if
# you want to run only the similarities checker, you can use "--disable=all
# --enable=similarities". If you want to run only the classes checker, but have
# no Warning level messages displayed, use "--disable=all --enable=classes
# --disable=W".
disable=raw-checker-failed,
        bad-inline-option,
        locally-disabled,
        file-ignored,
        suppressed-message,
        useless-suppression,
        deprecated-pragma,
        use-symbolic-message-instead

# Enable the message, report, category or checker with the given id(s). You can
# either give multiple identifier separated by comma (,) or put this option
# multiple time (only on the command line, not in the configuration file where
# it should appear only once). See also the "--disable" option for examples.
enable=c-extension-no-member


[REPORTS]

# Python expression which should return a score less than or equal to 10. You
# have access to the variables 'fatal', 'error', 'warning', 'refactor',
# 'convention', and 'info' which contain the number of messages in each
# category, as well as 'statement' which is the total number of statements
# analyzed. This score is used by the global evaluation report (RP0004).
evaluation=10.0 - ((float(5 * error + warning + refactor + convention) / statement) * 10)

# Template used to display messages. This is a python new-style format string
# used to format the message information. See doc for all details.
#msg-template=

# Set the output format. Available formats are text, parseable, colorized, json
# and msvs (visual studio). You can also give a reporter class, e.g.
# mypackage.mymodule.MyReporterClass.
output-format=text

# Tells whether to display a full report or only the messages.
reports=no

# Activate the evaluation score.
score=yes


[REFACTORING]

# Maximum number of nested blocks for function / method body
max-nested-blocks=5

# Complete name of functions that never returns. When checking for
# inconsistent-return-statements if a never returning function is called then
# it will be considered as an explicit return statement and no message will be
# printed.
never-returning-functions=sys.exit,argparse.parse_error


[LOGGING]

# The type of string formatting that logging methods do. `old` means using %
# formatting, `new` is for `{}` formatting.
logging-format-style=old

# Logging modules to check that the string format arguments are in logging
# function parameter format.
logging-modules=logging


[SPELLING]

# Limits count of emitted suggestions for spelling mistakes.
max-spelling-suggestions=4

# Spelling dictionary name. Available dictionaries: none. To make it work,
# install the 'python-enchant' package.
spelling-dict=

# List of comma separated words that should be considered directives if they
# appear and the beginning of a comment and should not be checked.
spelling-ignore-comment-directives=fmt: on,fmt: off,noqa:,noqa,nosec,isort:skip,mypy:

# List of comma separated words that should not be checked.
spelling-ignore-words=

# A path to a file that contains the private dictionary; one word per line.
spelling-private-dict-file=

# Tells whether to store unknown words to the private dictionary (see the
# --spelling-private-dict-file option) instead of raising a message.
spelling-store-unknown-words=no


[MISCELLANEOUS]

# List of note tags to take in consideration, separated by a comma.
notes=FIXME,
      XXX,
      TODO

# Regular expression of note tags to take in consideration.
#notes-rgx=


[TYPECHECK]

# List of decorators that produce context managers, such as
# contextlib.contextmanager. Add to this list to register other decorators that
# produce valid context managers.
contextmanager-decorators=contextlib.contextmanager

# List of members which are set dynamically and missed by pylint inference
# system, and so shouldn't trigger E1101 when accessed. Python regular
# expressions are accepted.
generated-members=

# Tells whether missing members accessed in mixin class should be ignored. A
# class is considered mixin if its name matches the mixin-class-rgx option.
ignore-mixin-members=yes

# Tells whether to warn about missing members when the owner of the attribute
# is inferred to be None.
ignore-none=yes

# This flag controls whether pylint should warn about no-member and similar
# checks whenever an opaque object is returned when inferring. The inference
# can return multiple potential results while evaluating a Python object, but
# some branches might not be evaluated, which results in partial inference. In
# that case, it might be useful to still emit no-member and other checks for
# the rest of the inferred objects.
ignore-on-opaque-inference=yes

# List of class names for which member attributes should not be checked (useful
# for classes with dynamically set attributes). This supports the use of
# qualified names.
ignored-classes=optparse.Values,thread._local,_thread._local

# List of module names for which member attributes should not be checked
# (useful for modules/projects where namespaces are manipulated during runtime
# and thus existing member attributes cannot be deduced by static analysis). It
# supports qualified module names, as well as Unix pattern matching.
ignored-modules=

# Show a hint with possible names when a member name was not found. The aspect
# of finding the hint is based on edit distance.
missing-member-hint=yes

# The minimum edit distance a name should have in order to be considered a
# similar match for a missing member name.
missing-member-hint-distance=1

# The total number of similar names that should be taken in consideration when
# showing a hint for a missing member.
missing-member-max-choices=1

# Regex pattern to define which classes are considered mixins ignore-mixin-
# members is set to 'yes'
mixin-class-rgx=.*[Mm]ixin

# List of decorators that change the signature of a decorated function.
signature-mutators=


[VARIABLES]

# List of additional names supposed to be defined in builtins. Remember that
# you should avoid defining new builtins when possible.
additional-builtins=

# Tells whether unused global variables should be treated as a violation.
allow-global-unused-variables=yes

# List of names allowed to shadow builtins
allowed-redefined-builtins=

# List of strings which can identify a callback function by name. A callback
# name must start or end with one of those strings.
callbacks=cb_,
          _cb

# A regular expression matching the name of dummy variables (i.e. expected to
# not be used).
dummy-variables-rgx=_+$|(_[a-zA-Z0-9_]*[a-zA-Z0-9]+?$)|dummy|^ignored_|^unused_

# Argument names that match this expression will be ignored. Default to name
# with leading underscore.
ignored-argument-names=_.*|^ignored_|^unused_

# Tells whether we should check for unused import in __init__ files.
init-import=no

# List of qualified module names which can have objects that can redefine
# builtins.
redefining-builtins-modules=six.moves,past.builtins,future.builtins,builtins,io


[FORMAT]

# Expected format of line ending, e.g. empty (any line ending), LF or CRLF.
expected-line-ending-format=

# Regexp for a line that is allowed to be longer than the limit.
ignore-long-lines=^\s*(# )?<?https?://\S+>?$

# Number of spaces of indent required inside a hanging or continued line.
indent-after-paren=4

# String used as indentation unit. This is usually "    " (4 spaces) or "\t" (1
# tab).
indent-string='    '

# Maximum number of characters on a single line.
max-line-length=100

# Maximum number of lines in a module.
max-module-lines=1000

# Allow the body of a class to be on the same line as the declaration if body
# contains single statement.
single-line-class-stmt=no

# Allow the body of an if to be on the same line as the test if there is no
# else.
single-line-if-stmt=no


[SIMILARITIES]

# Comments are removed from the similarity computation
ignore-comments=yes

# Docstrings are removed from the similarity computation
ignore-docstrings=yes

# Imports are removed from the similarity computation
ignore-imports=no

# Signatures are removed from the similarity computation
ignore-signatures=no

# Minimum lines number of a similarity.
min-similarity-lines=4


[STRING]

# This flag controls whether inconsistent-quotes generates a warning when the
# character used as a quote delimiter is used inconsistently within a module.
check-quote-consistency=no

# This flag controls whether the implicit-str-concat should generate a warning
# on implicit string concatenation in sequences defined over several lines.
check-str-concat-over-line-jumps=no


[BASIC]

# Naming style matching correct argument names.
argument-naming-style=snake_case

# Regular expression matching correct argument names. Overrides argument-
# naming-style. If left empty, argument names will be checked with the set
# naming style.
#argument-rgx=

# Naming style matching correct attribute names.
attr-naming-style=snake_case

# Regular expression matching correct attribute names. Overrides attr-naming-
# style. If left empty, attribute names will be checked with the set naming
# style.
#attr-rgx=

# Bad variable names which should always be refused, separated by a comma.
bad-names=foo,
          bar,
          baz,
          toto,
          tutu,
          tata

# Bad variable names regexes, separated by a comma. If names match any regex,
# they will always be refused
bad-names-rgxs=

# Naming style matching correct class attribute names.
class-attribute-naming-style=any

# Regular expression matching correct class attribute names. Overrides class-
# attribute-naming-style. If left empty, class attribute names will be checked
# with the set naming style.
#class-attribute-rgx=

# Naming style matching correct class constant names.
class-const-naming-style=UPPER_CASE

# Regular expression matching correct class constant names. Overrides class-
# const-naming-style. If left empty, class constant names will be checked with
# the set naming style.
#class-const-rgx=

# Naming style matching correct class names.
class-naming-style=PascalCase

# Regular expression matching correct class names. Overrides class-naming-
# style. If left empty, class names will be checked with the set naming style.
#class-rgx=

# Naming style matching correct constant names.
const-naming-style=UPPER_CASE

# Regular expression matching correct constant names. Overrides const-naming-
# style. If left empty, constant names will be checked with the set naming
# style.
#const-rgx=

# Minimum line length for functions/classes that require docstrings, shorter
# ones are exempt.
docstring-min-length=-1

# Naming style matching correct function names.
function-naming-style=snake_case

# Regular expression matching correct function names. Overrides function-
# naming-style. If left empty, function names will be checked with the set
# naming style.
#function-rgx=

# Good variable names which should always be accepted, separated by a comma.
good-names=i,
           j,
           k,
           ex,
           Run,
           _,
           df,
           rp

# Good variable names regexes, separated by a comma. If names match any regex,
# they will always be accepted
good-names-rgxs=

# Include a hint for the correct naming format with invalid-name.
include-naming-hint=no

# Naming style matching correct inline iteration names.
inlinevar-naming-style=any

# Regular expression matching correct inline iteration names. Overrides
# inlinevar-naming-style. If left empty, inline iteration names will be checked
# with the set naming style.
#inlinevar-rgx=

# Naming style matching correct method names.
method-naming-style=snake_case

# Regular expression matching correct method names. Overrides method-naming-
# style. If left empty, method names will be checked with the set naming style.
#method-rgx=

# Naming style matching correct module names.
module-naming-style=snake_case

# Regular expression matching correct module names. Overrides module-naming-
# style. If left empty, module names will be checked with the set naming style.
#module-rgx=

# Colon-delimited sets of names that determine each other's naming style when
# the name regexes allow several styles.
name-group=

# Regular expression which should only match function or class names that do
# not require a docstring.
no-docstring-rgx=^_

# List of decorators that produce properties, such as abc.abstractproperty. Add
# to this list to register other decorators that produce valid properties.
# These decorators are taken in consideration only for invalid-name.
property-classes=abc.abstractproperty

# Regular expression matching correct type variable names. If left empty, type
# variable names will be checked with the set naming style.
#typevar-rgx=

# Naming style matching correct variable names.
variable-naming-style=snake_case

# Regular expression matching correct variable names. Overrides variable-
# naming-style. If left empty, variable names will be checked with the set
# naming style.
#variable-rgx=


[CLASSES]

# Warn about protected attribute access inside special methods
check-protected-access-in-special-methods=no

# List of method names used to declare (i.e. assign) instance attributes.
defining-attr-methods=__init__,
                      __new__,
                      setUp,
                      __post_init__

# List of member names, which should be excluded from the protected access
# warning.
exclude-protected=_asdict,
                  _fields,
                  _replace,
                  _source,
                  _make

# List of valid names for the first argument in a class method.
valid-classmethod-first-arg=cls

# List of valid names for the first argument in a metaclass class method.
valid-metaclass-classmethod-first-arg=cls


[IMPORTS]

# List of modules that can be imported at any level, not just the top level
# one.
allow-any-import-level=

# Allow wildcard imports from modules that define __all__.
allow-wildcard-with-all=no

# Analyse import fallback blocks. This can be used to support both Python 2 and
# 3 compatible code, which means that the block might have code that exists
# only in one or another interpreter, leading to false positives when analysed.
analyse-fallback-blocks=no

# Deprecated modules which should not be used, separated by a comma.
deprecated-modules=

# Output a graph (.gv or any supported image format) of external dependencies
# to the given file (report RP0402 must not be disabled).
ext-import-graph=

# Output a graph (.gv or any supported image format) of all (i.e. internal and
# external) dependencies to the given file (report RP0402 must not be
# disabled).
import-graph=

# Output a graph (.gv or any supported image format) of internal dependencies
# to the given file (report RP0402 must not be disabled).
int-import-graph=

# Force import order to recognize a module as part of the standard
# compatibility libraries.
known-standard-library=

# Force import order to recognize a module as part of a third party library.
known-third-party=enchant

# Couples of modules and preferred modules, separated by a comma.
preferred-modules=


[DESIGN]

# List of regular expressions of class ancestor names to ignore when counting
# public methods (see R0903)
exclude-too-few-public-methods=

# List of qualified class names to ignore when counting class parents (see
# R0901)
ignored-parents=

# Maximum number of arguments for function / method.
max-args=5

# Maximum number of attributes for a class (see R0902).
max-attributes=30

# Maximum number of boolean expressions in an if statement (see R0916).
max-bool-expr=5

# Maximum number of branch for function / method body.
max-branches=12

# Maximum number of locals for function / method body.
max-locals=15

# Maximum number of parents for a class (see R0901).
max-parents=7

# Maximum number of public methods for a class (see R0904).
max-public-methods=20

# Maximum number of return / yield for function / method body.
max-returns=6

# Maximum number of statements in function / method body.
max-statements=50

# Minimum number of public methods for a class (see R0903).
min-public-methods=2


[EXCEPTIONS]

# Exceptions that will emit a warning when being caught. Defaults to
# "BaseException, Exception".
overgeneral-exceptions=BaseException,
                       Exception
```


# Runtime environment
We run all of our projects within google cloud run. All services expose an HTTP API that is called through [google cloud workflows](https://cloud.google.com/workflows). This allows us to build a completely serverless data pipeline with retries and predictable execution. 

The downside of using this approach is no individual API call can take more than 60 minutes, the maximum time for a single request in google cloud run. This has lead to generally positive results, it requires us to break things into smaller chunks that can fit within this constraint.

# Deployment

All services are deployed through our CI/CD infrastructure with a combination of github workflows and google cloud build. Here is an example of our deployment workflow file:

```yaml
name: <project-name> Deploy
on:
  push:
    branches: [main]
    paths:
      - "projects/<project-name>/**/*"
defaults:
  run:
    shell: bash
    working-directory: projects/<project-name>
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  Deploy:
    runs-on: ubuntu-latest
    timeout-minutes: 30
    steps:
      - uses: actions/checkout@v3
      - id: "auth"
        uses: "google-github-actions/auth@v0"
        with:
          project_id: "<gcloud-project-name>"
          credentials_json: "${{ secrets.GCP_SA_KEY }}"

      - name: "setup cloud SDK"
        uses: "google-github-actions/setup-gcloud@v0"

      - name: "gcloud build"
        run: |-
          gcloud builds submit \
          --timeout=1800 \
          --machine-type=e2-highcpu-8 \
          --gcs-log-dir=gs://<gcloud-project-name>_cloudbuild/<project-name> \
          --tag us-central1-docker.pkg.dev/<gcloud-project-name>/cloud-run-source-deploy/<project-name>

      - id: "deploy"
        uses: "google-github-actions/deploy-cloudrun@v0"
        with:
          service: "<service-name>"
          image: "us-central1-docker.pkg.dev/<gcloud-project-name>/cloud-run-source-deploy/<project-name>:latest"
          region: "us-central1"
          env_vars: |
            PYTHONUNBUFFERED=True,PROJECT_ID=<gcloud-project-name>,BUCKET_NAME=<gcloud-project-name>,AIRTABLE_TOKEN=${{ secrets.AIRTABLE_TOKEN }},GMAPS_API_KEY=${{ secrets.GMAPS_API_KEY }}
          timeout: 3600
          flags: --min-instances=0 --max-instances=5 --memory=16Gi --cpu=4
```
When we merge to main the deploy automatically kicks off and deploys it into our production environment.

An interesting option here:
```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```
This allows us to cancel deploys mid way if we merge a few things quickly. The last one is the only one deployed. This ends up saving a decent amount of money.


# Working in this environment as a Dev

## First time checkout
Download anaconda from https://www.anaconda.com/products/distribution

```bash
bash ~/Downloads/Anaconda<version>.sh
git clone <mono-repo>
```
```bash
cd <mono-repo>/projects/<project-name>
conda env create -f environment.yml
conda activate <project-name>
```
Now you have a working environment. As long as your have activated the env you can python main.py or in our case use uvicorn to start the app.

```bash
uvicorn main:app
```
Running in docker we use an entrypoint.sh file that looks like this:
```bash
#!/bin/bash

if [[ -z "${PORT}" ]]; then
  PORT="8080"
else
  PORT="${PORT}"
fi

source /venv/bin/activate && uvicorn main:app --host "0.0.0.0" --port ${PORT} --workers 1
```

Example docker file:
```bash
# The build-stage image:
FROM --platform=linux/amd64 condaforge/mambaforge:4.13.0-1 AS build

# Install the package as normal:
COPY environment.yml . 

RUN mamba env create -f environment.yml

# Install conda-pack:
RUN mamba install -c conda-forge conda-pack

# Use conda-pack to create a standalone enviornment
# in /venv:
RUN conda-pack -n <project-name> -o /tmp/env.tar && \
  mkdir /venv && cd /venv && tar xf /tmp/env.tar && \
  rm /tmp/env.tar

# We've put venv in same path it'll be in final image,
# so now fix up paths:
RUN /venv/bin/conda-unpack

# The prod image; we can use Debian as the
# base image since the Conda env also includes Python
# for us.
FROM --platform=linux/amd64 debian:buster-slim AS prod

# Copy /venv from the previous stage:
COPY --from=build /venv /venv

COPY . /app
WORKDIR /app

# When image is run, run the code with the environment
# activated:
SHELL ["/bin/bash", "-c"]

ENV PORT 8080

ENTRYPOINT [ "/app/entrypoint.sh" ]
```

If you want to use the docker file **MAKE SURE TO UPDATE THE PROJECT-NAME**.

# Benefits & Drawbacks

Overall I am pretty happy with this solution, it gives us a simple way to allow all engineers to quickly get working dev environments for every python project. Leveraging google cloud workflows and google cloud run has allowed us to move quickly and build a robust production ready data pipeline with very small teams. Anaconda is large and painful to build but with micromamba and miniconda we were able to reduce our CI times to the point of a small annoyance.

### Pros:
* Easy to teach new devs
* CI integration "Just Works" meaning it fades to the back and no one thinks about it
* Devs can run it locally or in their own cloud environment quickly
* Workflows allows us to not think about orchestration
* Extremely low costs, we have the majority of our services set to scale to 0, this means services aren't running unless actively being used.

### Cons:

* Large docker images, generally over 1gb (Anaconda)
* Limited to only 60 min execution time per API request
* Tied directly to google cloud





