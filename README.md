# Rated list DHT

### Quick Simulation
```bash
docker build --tag ratedlist-node .
docker compose up
```

**Expected Output:** As of 04.08.2024, you must see all the spawned nodes bootstrapping through a single node.

### ToDo

- [ ] improve the bootstrapping process by saving the keys of initial nodes and later using them. Make it a ceremony rather than a reproducible step
- [ ] implement the basic rated list spec
- [ ] research on libp2p metrics
- [ ] research on p2p integration tests
- [ ] write unit tests
