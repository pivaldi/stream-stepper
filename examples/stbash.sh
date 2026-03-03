#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR

source "$SCRIPT_DIR/st.bash"

st.h1 "Production Deployment Pipeline"
st.h2 "Phase 1: Environment Setup"

st.doing "Cleaning workspace"
st.do sleep 1
st.do echo "Removing old build artifacts from /tmp/build..."
st.done

st.doing "Fetching external dependencies"
st.do sleep 1.5
st.warn "Deprecated package detected (v1.0.2). Update recommended."

st.doing "Update modules"
st.nothingTodo

st.h2 "Phase 2: Build & Test"
st.h3 "Frontend Assets"
st.doing "Minifying CSS and JS files"
st.do sleep 2
st.done

st.h3 "Backend Binaries"
st.doing "Running Go unit tests"
st.do sleep 1
st.do echo "ok      github.com/pivaldi/app   0.452s"
st.do echo "ok      github.com/pivaldi/utils 0.112s"
st.done "PASSED"

st.doing "Generating code coverage report"
st.do sleep 1
st.skipped

st.h2 "Phase 3: Finalization"
st.doing "Pushing Docker image to registry"
st.do sleep 2
st.do echo "digest: sha256:7b5d1e4 size: 2314"
st.done "PUSHED"

st.success "Pipeline Finished Successfully"

st.doing "Post-deployment health check"
st.do sleep 1
st.fail "Server responded with 500 Internal Server Error"
