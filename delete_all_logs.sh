#!/bin/bash
#
# Go to the classic tokens page
# Go to: https://github.com/settings/tokens
#
# Do NOT click "Generate new token (fine-grained)"
#
# Click "Generate new token (classic)" (at the bottom of the page)
#
# Note: delete-runs
# Expiration: No expiration (or 90 days)
#
# Select the scopes:
# ☑️ repo (ALL - it will check all of them automatically)
#    ☑️ repo:status
#    ☑️ repo_deployment  
#    ☑️ public_repo
#    ☑️ repo:invite
#    ☑️ security_events
# ☑️ workflow  <-- ESSENTIAL!
# ☑️ admin:repo_hook
# ☑️ delete_repo (optional)
#

# Function to detect the repository from .git
detect_repo() {
    # Check if we're inside a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        echo "❌ Error: You are not inside a git repository."
        echo "Run this script inside a git repository."
        exit 1
    fi
    
    # Get the remote URL
    local url=$(git config --get remote.origin.url 2>/dev/null)
    
    # If there's no origin, try the first remote
    if [ -z "$url" ]; then
        local first_remote=$(git remote 2>/dev/null | head -n1)
        if [ -n "$first_remote" ]; then
            url=$(git config --get "remote.$first_remote.url")
        else
            echo "❌ Error: No remote configured in this repository."
            echo "Configure a remote with: git remote add origin <url>"
            exit 1
        fi
    fi
    
    # Extract user/repository
    local repo=""
    if [[ "$url" =~ git@[^:]+:(.+).git$ ]]; then
        repo="${BASH_REMATCH[1]}"
    elif [[ "$url" =~ https?://[^/]+/(.+).git$ ]]; then
        repo="${BASH_REMATCH[1]}"
    elif [[ "$url" =~ git@[^:]+:(.+)$ ]]; then
        repo="${BASH_REMATCH[1]}"
    elif [[ "$url" =~ https?://[^/]+/(.+)$ ]]; then
        repo="${BASH_REMATCH[1]}"
    elif [[ "$url" =~ ssh://[^/]+/(.+).git$ ]]; then
        repo="${BASH_REMATCH[1]}"
    else
        repo=$(echo "$url" | sed 's/.*github\.com[:/]//;s/\.git$//')
    fi
    
    repo=$(echo "$repo" | sed 's/\/$//')
    
    if [ -z "$repo" ]; then
        echo "❌ Error: Could not extract the repository name from the URL: $url"
        exit 1
    fi
    
    echo "$repo"
}

# Detect the current repository
REPO=$(detect_repo)

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║     🔥 MASS RUN DELETER                                  ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""
echo "📁 Repository: $REPO"
echo ""

# Check if the token was provided as an argument or environment variable
if [ -n "$1" ]; then
    TOKEN="$1"
elif [ -n "$GITHUB_TOKEN" ]; then
    TOKEN="$GITHUB_TOKEN"
else
    echo "❌ Error: No token provided!"
    echo ""
    echo "Usage:"
    echo "  $0 YOUR_TOKEN"
    echo ""
    echo "Or export it as an environment variable first:"
    echo "  export GITHUB_TOKEN=YOUR_TOKEN"
    echo "  $0"
    echo ""
    echo "🔑 Create a Classic token at: https://github.com/settings/tokens"
    echo "   Required scopes: repo, workflow"
    exit 1
fi

if [ -z "$TOKEN" ]; then
    echo "❌ Token not provided!"
    exit 1
fi

# Test the token
echo "🔍 Testing token..."
RESPONSE=$(curl -s -H "Authorization: token $TOKEN" \
    "https://api.github.com/repos/$REPO" 2>/dev/null)

if echo "$RESPONSE" | grep -q "Bad credentials"; then
    echo "❌ Invalid token! Make sure it is a CLASSIC token."
    echo "   Required scopes: repo, workflow"
    exit 1
fi

echo "✅ Valid token!"
echo ""

# Fetch ALL runs
echo "🔍 Fetching all runs..."
echo ""

PAGE=1
ALL_IDS=()

while true; do
    echo "  📄 Fetching page $PAGE..."
    
    RESPONSE=$(curl -s -H "Authorization: token $TOKEN" \
        "https://api.github.com/repos/$REPO/actions/runs?per_page=100&page=$PAGE" 2>/dev/null)
    
    # Check if the response is valid
    if echo "$RESPONSE" | grep -q "API rate limit"; then
        echo "  ⏳ Rate limit reached. Waiting 60 seconds..."
        sleep 60
        continue
    fi
    
    IDS=$(echo "$RESPONSE" | jq -r '.workflow_runs[].id' 2>/dev/null)
    
    if [ -z "$IDS" ] || [ "$IDS" = "null" ] || [ "$IDS" = "" ]; then
        break
    fi
    
    COUNT=0
    while IFS= read -r id; do
        if [ -n "$id" ] && [ "$id" != "null" ]; then
            ALL_IDS+=("$id")
            COUNT=$((COUNT + 1))
        fi
    done <<< "$IDS"
    
    echo "  ✅ Found $COUNT runs on this page"
    
    # Check if there are more pages
    TOTAL_COUNT=$(echo "$RESPONSE" | jq -r '.total_count' 2>/dev/null)
    if [ -n "$TOTAL_COUNT" ] && [ ${#ALL_IDS[@]} -ge "$TOTAL_COUNT" ]; then
        break
    fi
    
    PAGE=$((PAGE + 1))
    
    # Safety limit (avoids infinite loop)
    if [ $PAGE -gt 50 ]; then
        break
    fi
done

TOTAL=${#ALL_IDS[@]}

echo ""
echo "📊 TOTAL found: $TOTAL runs"
echo ""

if [ "$TOTAL" -eq 0 ]; then
    echo "✅ No runs found!"
    exit 0
fi

echo "📋 First 5 runs:"
for i in {0..4}; do
    if [ $i -lt $TOTAL ]; then
        echo "  Run ${ALL_IDS[$i]}"
    fi
done
echo ""

read -p "⚠️  Delete ALL $TOTAL runs? (type 'YES' to confirm): " CONFIRMACAO

if [ "$CONFIRMACAO" != "YES" ]; then
    echo "❌ Cancelled."
    exit 0
fi

echo ""
echo "🚀 Deleting in batches of 20..."
echo ""

SUCESSOS=0
FALHAS=0
JA_DELETADOS=0
CONTADOR=0

for id in "${ALL_IDS[@]}"; do
    CONTADOR=$((CONTADOR + 1))
    PERCENT=$((CONTADOR * 100 / TOTAL))
    
    echo -n "[$PERCENT%] $CONTADOR/$TOTAL - Deleting $id ... "
    
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE \
        -H "Authorization: token $TOKEN" \
        "https://api.github.com/repos/$REPO/actions/runs/$id" 2>/dev/null)
    
    if [ "$STATUS" = "204" ]; then
        echo "✅"
        SUCESSOS=$((SUCESSOS + 1))
    elif [ "$STATUS" = "404" ]; then
        echo "⏭️  (already deleted)"
        JA_DELETADOS=$((JA_DELETADOS + 1))
    else
        echo "❌ (HTTP $STATUS)"
        FALHAS=$((FALHAS + 1))
    fi
    
    # Pause to avoid overloading
    sleep 0.05
done

echo ""
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║                     📊 FINAL SUMMARY                     ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""
echo "  ✅ Successfully deleted: $SUCESSOS"
if [ $JA_DELETADOS -gt 0 ]; then
    echo "  ⏭️  Already deleted: $JA_DELETADOS"
fi
if [ $FALHAS -gt 0 ]; then
    echo "  ❌ Failures: $FALHAS"
fi
echo ""

# Check if there are still remaining runs
echo "🔍 Checking remaining runs..."
REMAINING=$(curl -s -H "Authorization: token $TOKEN" \
    "https://api.github.com/repos/$REPO/actions/runs?per_page=1" 2>/dev/null | \
    jq -r '.total_count' 2>/dev/null)

if [ -n "$REMAINING" ] && [ "$REMAINING" -gt 0 ]; then
    echo "⚠️  There are still $REMAINING runs (might be GitHub cache)"
    echo "   Wait 1 minute and run again"
else
    echo "🎉 MISSION ACCOMPLISHED! All runs have been deleted!"
fi

echo ""
echo "💡 Tip: To use in other repositories, run inside the project folder"
echo "   Or pass the token as an argument: $0 YOUR_TOKEN"