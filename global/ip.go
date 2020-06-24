package global

import (
    "net"
    "strings"
    "fmt"
)


// Get network interface index by it's IP.
// Usage example: getInterfaceIndexByIP(net.ParseIP("fe80::aaaa:bbbb:cccc:dddd"))
func getInterfaceIndexByIP(addr net.IP) (index int, e error) {
    if addr == nil {
        e = fmt.Errorf("IP address is not correct")
        return
    }

    iface, err := net.Interfaces()
    if e != nil {
        e = err
        return
    }

    for _, i := range iface {
        addrs, err := i.Addrs()
        if err != nil {
            e = err
            return
        }

        for _, a := range addrs {
            address := net.ParseIP(strings.SplitN(a.String(), "/", 2)[0])
            if addr.Equal(address) {
                index = i.Index
                return
            }
        }
    }

    e = fmt.Errorf("Could not found interface by this IP address: '%s'", addr.String())

    return
}


// check if an address is IPv6 link local address (fe80::/10)
func isIPv6LinkLocalAddress(addr net.IP) bool {
    if addr.To4() != nil {
        return false
    }
    return addr.IsLinkLocalUnicast()
}


// get all IPs from command-line option -bind-ip
func getIPs(bindIP string) (ips []string, err error) {
    anyIP := []string{"0.0.0.0"}

    if bindIP == "" {
        ips = anyIP
        return
    }

    var invalidIPs []string
    allIPs := make(map[string]interface{})

    for _, item := range strings.Split(bindIP, ",") {
        item = strings.TrimSpace(item)
        if item == "" {
            continue
        }
        add := net.ParseIP(item)
        if add != nil {
            if add.To4() != nil {
                // IPv4 address
                allIPs[add.String()] = nil
            } else {
                // IPv6 address

                if isIPv6LinkLocalAddress(add) {
                    // if the address is an IPv6 link local address (fe80::/10),
                    // add a scope (interface index) to the end of the address.
                    index, e := getInterfaceIndexByIP(add)
                    if e != nil {
                        err = e
                        return
                    }
                    allIPs[fmt.Sprintf("[%s%%%d]", add.String(), index)] = nil
                } else {
                    allIPs[fmt.Sprintf("[%s]", add.String())] = nil
                }
            }
        } else {
            invalidIPs = append(invalidIPs, item)
        }
    }

    if len(invalidIPs) > 0 {
        err = fmt.Errorf("Invalid IP: %s", strings.Join(invalidIPs, ", "))
        return
    }


    if _, ok := allIPs["0.0.0.0"]; ok {
        ips = anyIP
        return
    }
    if _, ok := allIPs["[::]"]; ok {
        ips = anyIP
        return
    }

    for key, _ := range allIPs {
        ips = append(ips, key)
    }

    if len(ips) == 0 {
        err = fmt.Errorf("No IP address provided")
        return
    }

    return
}

