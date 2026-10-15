# Agent Language and Quality Guidelines

1. **English Language Standard**:
   - All generated code, comments, user interface labels, buttons, navigation items, modals, and commit messages MUST be written in English.
   - Use standard X (Twitter) terminology: "Home", "Explore", "Notifications", "Messages", "Grok", "Profile", "Post", "Drafts", "What is happening?!", "Who to follow", "What's happening", "For you", "Following", "Subscribe", etc.

2. **Graphify Auto-Update Rule**:
   - Whenever you write new code, modify existing code, or add new features, proactively rebuild the Graphify architecture graph:
     - Run `graphify . --code-only` in the workspace root.
     - Run `graphify cluster-only .` in the workspace root.
   - Always inform the user that the graph has been updated.

3. **Docker Build Rule**:
   - Always test and build with Docker (`docker-compose.prod.yml`) ensuring clean container recreation without stale IP caches.
