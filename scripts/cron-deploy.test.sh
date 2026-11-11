#!/usr/bin/env bash
set -euo pipefail

SCRIPT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/cron-deploy"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

passed=0
failed=0

ok() {
    passed=$((passed + 1))
    printf 'ok   %s\n' "$1"
}

bad() {
    failed=$((failed + 1))
    printf 'FAIL %s\n' "$1"
}

check() {
    local name="$1"
    shift
    if "$@"; then ok "$name"; else bad "$name"; fi
}

contains() {
    grep -qF -- "$2" "$1"
}

lacks() {
    ! grep -qF -- "$2" "$1" 2>/dev/null
}

mkdir -p "$WORK/shims"
cat >"$WORK/shims/docker" <<'EOF'
#!/usr/bin/env bash
printf 'docker %s\n' "$*" >>"$STUB_LOG"
if [ "$1" = "compose" ]; then
    for arg in "$@"; do
        if [ "$arg" = "up" ]; then
            printf 'up at %s\n' "$(git rev-parse HEAD)" >>"$STUB_LOG"
        fi
    done
fi
exit 0
EOF
cat >"$WORK/shims/curl" <<'EOF'
#!/usr/bin/env bash
for arg in "$@"; do
    if [ "$arg" = "-K" ]; then
        cfg="$(cat)"
        data=""
        all="$*"
        while [ $# -gt 0 ]; do
            if [ "$1" = "--data" ]; then data="$2"; fi
            shift
        done
        set -- $all
        printf 'slack %s\n' "$data" >>"$STUB_LOG"
        printf 'curl-args %s\n' "$*" >>"$STUB_LOG"
        printf 'slack-config %s\n' "$cfg" >>"$STUB_LOG"
        exit 0
    fi
done
printf 'health at %s\n' "$(git rev-parse HEAD)" >>"$STUB_LOG"
if [ "$(git rev-parse HEAD)" = "${STUB_BAD_SHA:-none}" ]; then
    exit 22
fi
exit 0
EOF
chmod +x "$WORK/shims/docker" "$WORK/shims/curl"

new_case() {
    CASE="$WORK/$1"
    mkdir -p "$CASE"
    git init -q --bare -b main "$CASE/origin.git"
    git clone -q "$CASE/origin.git" "$CASE/repo" 2>/dev/null
    (
        cd "$CASE/repo"
        git config user.email test@example.com
        git config user.name test
        git checkout -q -b main
        echo one >file
        printf 'services: {}\n' >compose.yaml
        printf 'services: {}\n' >compose.pi.yaml
        git add -A
        git commit -q -m one
        git push -q -u origin main
    )
    export STUB_LOG="$CASE/stub.log"
    : >"$STUB_LOG"
    unset STUB_BAD_SHA
    export PATH="$WORK/shims:$ORIG_PATH"
    export DEPLOY_ROOT="$CASE/repo"
    export HEALTH_ATTEMPTS=2
    export HEALTH_INTERVAL=0
    export SLACK_DM_URL="https://hooks.example.invalid/secret-token"
    export PORT=9942
    OUT="$CASE/out.log"
}

push_commit() {
    local other="$CASE/other"
    rm -rf "$other"
    git clone -q "$CASE/origin.git" "$other" 2>/dev/null
    (
        cd "$other"
        git config user.email test@example.com
        git config user.name test
        echo "$1" >file
        git commit -q -am "$1"
        git push -q origin main
    )
    git -C "$other" rev-parse HEAD
}

run() {
    set +e
    "$SCRIPT" >"$OUT" 2>&1
    RC=$?
    set -e
}

ORIG_PATH="$PATH"

new_case nochange
run
check "no change exits zero" test "$RC" -eq 0
check "no change prints nothing" test ! -s "$OUT"
check "no change does not touch docker" lacks "$STUB_LOG" "docker compose"
check "no change sends no slack message" lacks "$STUB_LOG" "slack"

new_case newcommit
sha="$(push_commit two)"
run
check "new commit exits zero" test "$RC" -eq 0
check "new commit moves HEAD to origin main" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$sha"
check "new commit builds with the pi overlay" contains "$STUB_LOG" "docker compose -f compose.yaml -f compose.pi.yaml --progress quiet build"
check "new commit starts the stack" contains "$STUB_LOG" "up at $sha"
check "new commit posts a deployed line" contains "$STUB_LOG" "Deployed ${sha:0:7}"
check "new commit never prints the slack url" lacks "$OUT" "secret-token"
check "new commit keeps the slack url out of curl arguments" test "$(grep '^curl-args ' "$STUB_LOG" | grep -c secret-token)" -eq 0
check "new commit passes the slack url on stdin" contains "$STUB_LOG" "slack-config url = \"https://hooks.example.invalid/secret-token\""

new_case dirty
sha="$(push_commit two)"
before="$(git -C "$CASE/repo" rev-parse HEAD)"
echo local >>"$CASE/repo/file"
run
check "dirty tree exits nonzero" test "$RC" -ne 0
check "dirty tree keeps HEAD" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$before"
check "dirty tree keeps local edits" contains "$CASE/repo/file" "local"
check "dirty tree does not rebuild" lacks "$STUB_LOG" "docker compose"
check "dirty tree alerts" contains "$STUB_LOG" "skipped"
run
check "dirty tree alerts only once per commit" test "$(grep -c '^slack ' "$STUB_LOG")" -eq 1

new_case rollback
before="$(git -C "$CASE/repo" rev-parse HEAD)"
sha="$(push_commit two)"
export STUB_BAD_SHA="$sha"
run
check "failed health exits nonzero" test "$RC" -ne 0
check "failed health rolls HEAD back" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$before"
check "failed health rebuilds the previous commit" contains "$STUB_LOG" "up at $before"
check "failed health posts a rollback line" contains "$STUB_LOG" "rolled back to ${before:0:7}"
n="$(grep -c '^docker compose.* build' "$STUB_LOG")"
run
check "a failed commit is not retried" test "$(grep -c '^docker compose.* build' "$STUB_LOG")" -eq "$n"
sha3="$(push_commit three)"
unset STUB_BAD_SHA
run
check "a later good commit deploys" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$sha3"
check "a good deploy clears the failure marker" test ! -e "$CASE/repo/.git/cron-deploy-failed"

new_case otherbranch
sha="$(push_commit two)"
git -C "$CASE/repo" checkout -q -b scratch
run
check "a checkout on another branch is returned to main" test "$(git -C "$CASE/repo" branch --show-current)" = "main"
check "a checkout on another branch deploys origin main" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$sha"

new_case locked
sha="$(push_commit two)"
before="$(git -C "$CASE/repo" rev-parse HEAD)"
mkdir "$CASE/repo/.git/cron-deploy.lock"
echo $$ >"$CASE/repo/.git/cron-deploy.lock/pid"
run
check "concurrent run exits zero" test "$RC" -eq 0
check "concurrent run does nothing" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$before"
check "concurrent run does not touch docker" lacks "$STUB_LOG" "docker compose"
check "concurrent run leaves the holder lock" test -d "$CASE/repo/.git/cron-deploy.lock"
rm -rf "$CASE/repo/.git/cron-deploy.lock"
mkdir "$CASE/repo/.git/cron-deploy.lock"
echo 999999 >"$CASE/repo/.git/cron-deploy.lock/pid"
run
check "a stale lock is reclaimed and the deploy runs" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$sha"
check "the lock is released after a run" test ! -d "$CASE/repo/.git/cron-deploy.lock"

printf '\n%s passed, %s failed\n' "$passed" "$failed"
test "$failed" -eq 0
