# Trail's Coffee Implementation Testing Guide

## Overview
This guide covers testing the complete Trail's Coffee three-tier WiFi system with Bitcoin payments.

## Test Environment Setup

### Hardware Requirements
- GL.iNet MT6000 router (or compatible OpenWrt device)
- Test devices (phones, laptops) for connecting to WiFi
- Bitcoin wallet with Lightning support (e.g., Phoenix, Muun)
- Cashu wallet (e.g., NutWallet, Minibits)

### Software Setup
1. **Build the package:**
   ```bash
   # For MT6000
   ./compile-to-router.sh [router-ip] --device=gl-mt6000

   # Or build .ipk and install via LuCI
   ```

2. **Configure pricing:**
   Edit `/etc/tollgate/config.json`:
   ```json
   {
     "step_size": 3600000,
     "accepted_mints": [{
       "price_per_step": 10,
       "price_unit": "hour"
     }]
   }
   ```

3. **Set Lightning address:**
   Edit `/etc/tollgate/identities.json`:
   ```json
   {
     "public_identities": [{
       "name": "owner",
       "lightning_address": "your-lightning@address.com"
     }]
   }
   ```

## Test Cases

### 1. WiFi Network Discovery
**Objective:** Verify all three networks are broadcast correctly

**Steps:**
1. Power on router after installation
2. Scan for WiFi networks on test device
3. Verify presence of:
   - `TrailsCoffee-Free-2.4GHz`
   - `TrailsCoffee-Free-5GHz`
   - `TrailsCoffee-Premium-2.4GHz`
   - `TrailsCoffee-Premium-5GHz`
   - `TrailsCoffee-Staff`

**Expected Results:**
- All networks visible
- Free networks show as open
- Premium networks show as open (but require payment)
- Staff network shows as secured (WPA3)

### 2. Free WiFi Tier Testing
**Objective:** Verify free tier functionality with bandwidth limiting

**Steps:**
1. Connect to `TrailsCoffee-Free-2.4GHz`
2. Verify captive portal appears
3. Click "Free WiFi" option
4. Attempt to access internet

**Expected Results:**
- Captive portal displays Trail's Coffee branding
- Free connection works
- Speed limited to ~2Mbps (test with speedtest.net)
- Session times out appropriately

### 3. Premium Payment Testing
**Objective:** Verify premium tier payment and unlimited access

**Steps:**
1. Connect to `TrailsCoffee-Premium-2.4GHz`
2. Verify portal shows payment options
3. Test Lightning payment:
   - Generate 10 sat invoice
   - Pay from Lightning wallet
   - Verify access granted
4. Test Cashu payment:
   - Generate Cashu token
   - Paste token in portal
   - Verify access granted

**Expected Results:**
- QR codes display correctly
- Payment processing works
- Unlimited speed access granted
- 1-hour session timer functions

### 4. Staff Network Testing
**Objective:** Verify password-protected staff network

**Steps:**
1. Attempt to connect to `TrailsCoffee-Staff`
2. Use default password: `TrailCoffee2024!`
3. Verify internet access without portal

**Expected Results:**
- Password authentication required
- Direct internet access (no captive portal)
- Unlimited speed
- Separate from customer networks

### 5. Bandwidth Limiting Verification
**Objective:** Confirm free tier is properly limited

**Test Tools:**
```bash
# On router, monitor traffic:
tc -s class show dev br-lan

# Speed test from client:
curl -o /dev/null http://speedtest.tele2.net/10MB.zip
```

**Expected Results:**
- Free tier limited to ~2Mbps
- Premium tier unlimited
- Staff network unlimited

### 6. Branding and UX Testing
**Objective:** Verify Trail's Coffee branding and user experience

**Steps:**
1. Connect to any customer network
2. Review portal appearance
3. Test responsive design on mobile
4. Verify Bitcoin education content
5. Test payment flows

**Expected Results:**
- Trail's Coffee logo displays
- Coffee shop messaging throughout
- Bitcoin education visible
- Mobile-friendly interface
- Clear payment options

### 7. Security Testing
**Objective:** Ensure network isolation and security

**Steps:**
1. Verify firewall rules isolate networks
2. Test that free/premium can't access staff network
3. Confirm HTTPS enforcement where needed
4. Check for open ports on WAN interface

**Expected Results:**
- Network segmentation works
- No unauthorized access between zones
- Secure payment processing

### 8. Load Testing
**Objective:** Verify system handles multiple users

**Steps:**
1. Connect 5-10 devices across different tiers
2. Generate simultaneous payments
3. Monitor router performance
4. Test session management

**Expected Results:**
- Router handles load without issues
- Payment processing scales
- Session isolation maintained

## Troubleshooting

### Common Issues

**WiFi Networks Not Appearing:**
```bash
# Check wireless config
uci show wireless
service network reload
```

**Captive Portal Not Loading:**
```bash
# Check NoDogSplash
service nodogsplash status
logread | grep nodogsplash
```

**Payments Not Processing:**
```bash
# Check TollGate service
service tollgate-wrt status
logread | grep tollgate
```

**Bandwidth Limiting Not Working:**
```bash
# Check traffic control
tc qdisc show dev br-lan
tc class show dev br-lan
```

### Debug Commands

```bash
# View all logs
logread | tail -50

# Check network interfaces
ifconfig
iwinfo

# Test internet connectivity
ping -c 3 1.1.1.1

# Check TollGate API
curl http://localhost:2121
```

## Performance Benchmarks

### Expected Performance
- **Boot Time:** < 2 minutes
- **Portal Load Time:** < 3 seconds
- **Payment Processing:** < 10 seconds
- **Concurrent Users:** 20-50 (depending on router specs)

### Resource Usage
- **CPU:** < 30% during normal operation
- **Memory:** < 100MB for base system
- **Storage:** < 50MB for portal assets

## Final Validation Checklist

- [ ] All three WiFi networks broadcasting
- [ ] Free tier bandwidth limited to 2Mbps
- [ ] Premium tier requires 10 sat payment
- [ ] Staff network password-protected
- [ ] Trail's Coffee branding throughout
- [ ] Bitcoin education content accessible
- [ ] QR codes for both payment methods
- [ ] Payment processing functional
- [ ] Session management working
- [ ] Network isolation maintained
- [ ] Mobile responsive design
- [ ] Error handling graceful

## Deployment Ready

Once all tests pass, the system is ready for Trail's Coffee deployment! 🎉
