#!/bin/bash

# Cleanup script to revert the artifact demo changes

echo "=========================================="
echo "Cleaning Up Artifact Demo"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}This will:${NC}"
echo "  1. Remove the failing test file"
echo "  2. Revert CI workflow to original triggers"
echo "  3. Remove demo documentation files"
echo "  4. Commit and push the cleanup"
echo ""
read -p "Continue? (y/N) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cleanup cancelled."
    exit 1
fi

echo ""
echo -e "${GREEN}Step 1: Removing demo files...${NC}"

# Remove demo test file
if [ -f "internal/metrics/artifact_test_demo.go" ]; then
    git rm internal/metrics/artifact_test_demo.go
    echo "  ✓ Removed artifact_test_demo.go"
else
    echo "  - artifact_test_demo.go not found (already removed?)"
fi

echo ""
echo -e "${GREEN}Step 2: Reverting CI workflow...${NC}"

# Revert the CI workflow trigger
cat > .github/workflows/ci.yml.tmp << 'EOF'
name: CI

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]
EOF

# Get the rest of the file after the trigger section
tail -n +10 .github/workflows/ci.yml >> .github/workflows/ci.yml.tmp
mv .github/workflows/ci.yml.tmp .github/workflows/ci.yml

echo "  ✓ Reverted .github/workflows/ci.yml to original triggers"

echo ""
echo -e "${GREEN}Step 3: Removing demo scripts...${NC}"

# Remove demo scripts
if [ -f "check_ci_artifacts.sh" ]; then
    rm check_ci_artifacts.sh
    echo "  ✓ Removed check_ci_artifacts.sh"
fi

if [ -f "cleanup_demo.sh" ]; then
    echo "  - Keeping cleanup_demo.sh (will be removed with git)"
fi

echo ""
echo -e "${GREEN}Step 4: Reviewing changes...${NC}"
git status

echo ""
echo -e "${YELLOW}Ready to commit cleanup?${NC}"
read -p "Commit and push? (y/N) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Changes staged but not committed."
    echo "Run 'git commit' manually when ready."
    exit 0
fi

echo ""
echo -e "${GREEN}Step 5: Committing cleanup...${NC}"

git add .github/workflows/ci.yml
git commit -m "Cleanup: Revert artifact demo changes

Removes:
- Intentional failing test (artifact_test_demo.go)
- Temporary branch trigger from CI workflow
- Demo scripts and documentation

CI should now pass cleanly on this branch.
Tests will only create artifacts on genuine failures."

echo ""
echo -e "${GREEN}Step 6: Pushing to remote...${NC}"
git push

echo ""
echo "=========================================="
echo -e "${GREEN}✅ Cleanup Complete!${NC}"
echo "=========================================="
echo ""
echo "Changes reverted:"
echo "  ✓ Demo test removed"
echo "  ✓ CI workflow restored"
echo "  ✓ Temporary files cleaned"
echo ""
echo "CI will now:"
echo "  - Run only on main/master pushes and PRs"
echo "  - Pass all tests (no intentional failures)"
echo "  - Create artifacts ONLY on real failures"
echo ""
echo "The artifact upload feature is still active!"
echo "It will work automatically when any real failure occurs."
echo ""
