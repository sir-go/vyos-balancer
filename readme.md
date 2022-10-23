## CLI tool for traffic balancing via [VyOS Router](https://vyos.io/vyos-router)

### Build

```bash
go mod download && go build -o vyosctl ./cmd/vyosctl
```
### Flags
`-c` <config file path> - path to `*.toml` config file
### Config

```
overflow_count = 6        # uplink overflow count
check_period = "5s"       # how often check uplinks

[vyos]
addr = "https://---"      # url to VyOS API
key  = "//"               # VyOS API token

[beeline]                 # 1st uplink name
alias = "enp3s0.304"      # interface
lz = 10                   # Mbps - zero level for traffic
l0 = 850                  # Mbps - lower edge
l1 = 2900                 # Mbps - upper edge
nat = "185.46.0.0/24"   # NAT external subnet

[TTK]                 # 2nd uplink name
...
```
