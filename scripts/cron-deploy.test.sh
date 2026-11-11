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
        if [ "$arg" = "build" ] && [ -s "$STUB_BUILD_FAILS" ] && [ "$(cat "$STUB_BUILD_FAILS")" -gt 0 ]; then
            echo $(($(cat "$STUB_BUILD_FAILS") - 1)) >"$STUB_BUILD_FAILS"
            printf 'build failed\n' >>"$STUB_LOG"
            exit 1
        fi
        if [ "$arg" = "up" ] && [ -s "$STUB_UP_FAILS" ] && [ "$(cat "$STUB_UP_FAILS")" -gt 0 ]; then
            echo $(($(cat "$STUB_UP_FAILS") - 1)) >"$STUB_UP_FAILS"
            printf 'up failed\n' >>"$STUB_LOG"
            exit 1
        fi
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
case " ${STUB_BAD_SHA:-none} " in *" $(git rev-parse HEAD) "*)
    exit 22;;
esac
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
        printf '.env\n' >.gitignore
        printf 'services: {}\n' >compose.yaml
        printf 'services: {}\n' >compose.pi.yaml
        git add -A
        git commit -q -m one
        git push -q -u origin main
        printf 'PORT=9942\n' >.env
    )
    export STUB_LOG="$CASE/stub.log"
    export STUB_BUILD_FAILS="$CASE/build-fails"
    : >"$STUB_LOG"
    echo 0 >"$STUB_BUILD_FAILS"
    export STUB_UP_FAILS="$CASE/up-fails"
    echo 0 >"$STUB_UP_FAILS"
    export DEPLOY_NOW=1000000
    export DEPLOY_NO_FLOCK=1
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
check "new commit runs up with the overlay files, detached, removing orphans" contains "$STUB_LOG" "docker compose -f compose.yaml -f compose.pi.yaml --progress quiet up -d --remove-orphans"
check "new commit builds before it starts" test "$(grep -n -m1 ' build$' "$STUB_LOG" | cut -d: -f1)" -lt "$(grep -n -m1 ' up -d' "$STUB_LOG" | cut -d: -f1)"
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
check "failed health says how to retry with the marker path" contains "$STUB_LOG" "rm $CASE/repo/.git/cron-deploy-failed"
check "failed health builds twice with the overlay files" test "$(grep -c '^docker compose -f compose.yaml -f compose.pi.yaml --progress quiet build$' "$STUB_LOG")" -eq 2
check "failed health runs up twice with the full argument list" test "$(grep -c '^docker compose -f compose.yaml -f compose.pi.yaml --progress quiet up -d --remove-orphans$' "$STUB_LOG")" -eq 2
n="$(grep -c '^docker compose.* build' "$STUB_LOG")"
run
check "a failed commit is not retried" test "$(grep -c '^docker compose.* build' "$STUB_LOG")" -eq "$n"
sha3="$(push_commit three)"
unset STUB_BAD_SHA
run
check "a later good commit deploys" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$sha3"
check "a good deploy clears the failure marker" test ! -e "$CASE/repo/.git/cron-deploy-failed"

new_case buildretry
before="$(git -C "$CASE/repo" rev-parse HEAD)"
sha="$(push_commit two)"
echo 2 >"$STUB_BUILD_FAILS"
run
check "a failed build exits nonzero" test "$RC" -ne 0
check "a failed build leaves HEAD on the running commit" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$before"
check "a failed build never starts the stack" lacks "$STUB_LOG" " up -d"
check "a failed build does not mark the commit failed" test ! -e "$CASE/repo/.git/cron-deploy-failed"
check "the first failed build alerts" contains "$STUB_LOG" "transient failure (build failed), attempt 1"
run
check "a retry inside the backoff window does nothing" test "$(grep -c '^build failed' "$STUB_LOG")" -eq 1
export DEPLOY_NOW=$((DEPLOY_NOW + 901))
run
check "a retry after the backoff window builds again" test "$(grep -c '^build failed' "$STUB_LOG")" -eq 2
check "the second failed build does not alert" test "$(grep -c '^slack ' "$STUB_LOG")" -eq 1
export DEPLOY_NOW=$((DEPLOY_NOW + 901))
run
check "the third attempt deploys the commit" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$sha"
check "a successful deploy clears the retry marker" test ! -e "$CASE/repo/.git/cron-deploy-retry"

new_case buildgiveup
sha="$(push_commit two)"
echo 99 >"$STUB_BUILD_FAILS"
export MAX_ATTEMPTS=3
for _ in 1 2 3; do
    run
    export DEPLOY_NOW=$((DEPLOY_NOW + 901))
done
run
check "giving up stops building after the maximum attempts" test "$(grep -c '^build failed' "$STUB_LOG")" -eq 3
check "giving up says how to retry with the marker path" contains "$STUB_LOG" "*gave up* after 3 attempts (build failed). To retry: rm $CASE/repo/.git/cron-deploy-retry"
unset MAX_ATTEMPTS

new_case startretry
before="$(git -C "$CASE/repo" rev-parse HEAD)"
sha="$(push_commit two)"
echo 1 >"$STUB_UP_FAILS"
run
check "a failed container start exits nonzero" test "$RC" -ne 0
check "a failed container start rolls the tree back" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$before"
check "a failed container start restarts the previous commit" contains "$STUB_LOG" "up at $before"
check "a failed container start is not marked bad" test ! -e "$CASE/repo/.git/cron-deploy-failed"
check "a failed container start is counted as a retry" contains "$CASE/repo/.git/cron-deploy-retry" "$sha 1"
export DEPLOY_NOW=$((DEPLOY_NOW + 901))
run
check "a failed container start is retried and deploys" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$sha"

new_case startdown
before="$(git -C "$CASE/repo" rev-parse HEAD)"
sha="$(push_commit two)"
echo 2 >"$STUB_UP_FAILS"
run
check "a failed start with a failed rollback exits nonzero" test "$RC" -ne 0
check "a failed start with a failed rollback records the retry state" contains "$CASE/repo/.git/cron-deploy-retry" "$sha 1"
check "a failed start with a failed rollback says the site is down" contains "$STUB_LOG" "site is *down*"
check "a failed start with a failed rollback says what to check" contains "$STUB_LOG" "docker compose ps"
slacks="$(grep -c '^slack ' "$STUB_LOG")"
builds="$(grep -c ' build$' "$STUB_LOG")"
DEPLOY_NOW=$((DEPLOY_NOW + 60))
run
check "the next run inside the backoff window does nothing" test "$(grep -c ' build$' "$STUB_LOG")" -eq "$builds"
check "the next run inside the backoff window sends nothing" test "$(grep -c '^slack ' "$STUB_LOG")" -eq "$slacks"
check "the next run inside the backoff window exits zero" test "$RC" -eq 0

new_case healthdown
before="$(git -C "$CASE/repo" rev-parse HEAD)"
sha="$(push_commit two)"
export STUB_BAD_SHA="$sha $before"
run
check "a failed health check with a failed rollback exits nonzero" test "$RC" -ne 0
check "a failed health check with a failed rollback says the site is down" contains "$STUB_LOG" "its health check and the site is *down*"
check "a failed health check with a failed rollback marks the commit bad" contains "$CASE/repo/.git/cron-deploy-failed" "$sha"
slacks="$(grep -c '^slack ' "$STUB_LOG")"
DEPLOY_NOW=$((DEPLOY_NOW + 60))
run
check "the next run after a down site stays quiet" test "$(grep -c '^slack ' "$STUB_LOG")" -eq "$slacks"

new_case fetchdown
sha="$(push_commit two)"
git -C "$CASE/repo" remote set-url origin "$CASE/missing.git"
run
check "an unreachable origin exits nonzero" test "$RC" -ne 0
run
check "an unreachable origin alerts only once" test "$(grep -c '^slack ' "$STUB_LOG")" -eq 1
git -C "$CASE/repo" remote set-url origin "$CASE/origin.git"
run
check "a reachable origin deploys again" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$sha"
check "a reachable origin clears the fetch alert marker" test ! -e "$CASE/repo/.git/cron-deploy-fetch"

new_case noenv
sha="$(push_commit two)"
before="$(git -C "$CASE/repo" rev-parse HEAD)"
rm "$CASE/repo/.env"
run
check "a missing .env exits nonzero" test "$RC" -ne 0
check "a missing .env does not deploy" lacks "$STUB_LOG" "docker compose"
check "a missing .env keeps HEAD" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$before"
check "a missing .env alerts" contains "$STUB_LOG" ".env is missing"

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
echo "$DEPLOY_NOW" >"$CASE/repo/.git/cron-deploy.lock/started"
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

new_case oldlock
sha="$(push_commit two)"
mkdir "$CASE/repo/.git/cron-deploy.lock"
echo $$ >"$CASE/repo/.git/cron-deploy.lock/pid"
echo $((DEPLOY_NOW - 7201)) >"$CASE/repo/.git/cron-deploy.lock/started"
run
check "a lock held by a live pid for over two hours is taken over" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$sha"
check "taking over an old lock alerts" contains "$STUB_LOG" "took over a stale lock"

if command -v flock >/dev/null 2>&1; then
    new_case flocked
    unset DEPLOY_NO_FLOCK
    sha="$(push_commit two)"
    before="$(git -C "$CASE/repo" rev-parse HEAD)"
    (
        exec 8>"$CASE/repo/.git/cron-deploy.flock"
        flock -n 8
        sleep 3
    ) &
    holder=$!
    sleep 1
    run
    check "flock: a run while the lock is held does nothing" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$before"
    wait "$holder"
    run
    check "flock: a leftover unlocked lock file does not block a deploy" test "$(git -C "$CASE/repo" rev-parse HEAD)" = "$sha"
    check "flock: the lock file stays but is not held" flock -n "$CASE/repo/.git/cron-deploy.flock" true
fi

printf '\n%s passed, %s failed\n' "$passed" "$failed"
test "$failed" -eq 0
