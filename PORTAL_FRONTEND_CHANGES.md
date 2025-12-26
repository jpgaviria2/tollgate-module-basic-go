# Trail's Coffee Portal Frontend Changes

## Overview
The captive portal React app needs modifications to support Trail's Coffee branding and payment flow.

## Required Changes

### 1. Payment Options Display
**File:** `src/components/PaymentOptions.js` (or equivalent)

**Current:** Shows separate tabs for Cashu and Lightning
**Target:** Show both payment methods prominently with QR codes

```javascript
// New payment options layout
<div className="payment-options">
  <h2>Choose Your Payment Method</h2>

  <div className="payment-grid">
    <div className="payment-option">
      <h3>⚡ Lightning (Instant)</h3>
      <div className="qr-code">
        {/* Lightning invoice QR code */}
      </div>
      <p>Pay 10 sats for 1 hour unlimited WiFi</p>
      <button>Copy Lightning Invoice</button>
    </div>

    <div className="payment-option">
      <h3>🥜 Cashu Ecash (Private)</h3>
      <div className="qr-code">
        {/* Cashu token QR code */}
      </div>
      <p>Pay with digital cash tokens</p>
      <button>Paste Cashu Token</button>
    </div>
  </div>
</div>
```

### 2. Branding Updates
**File:** `src/components/Header.js`

```javascript
// Update logo and branding
<div className="header">
  <img src="/trails-coffee-logo.svg" alt="Trail's Coffee" />
  <h1>Welcome to Trail's Coffee</h1>
  <p>Bitcoin-powered WiFi • 10 sats/hour unlimited access</p>
</div>
```

### 3. Educational Content
**File:** `src/components/BitcoinEducation.js` (new component)

```javascript
function BitcoinEducation() {
  return (
    <div className="bitcoin-education">
      <h3>Why Bitcoin at Trail's Coffee?</h3>
      <div className="benefits">
        <div className="benefit">
          <span className="icon">⚡</span>
          <span>Instant payments</span>
        </div>
        <div className="benefit">
          <span className="icon">🌐</span>
          <span>No intermediaries</span>
        </div>
        <div className="benefit">
          <span className="icon">🔒</span>
          <span>Financial privacy</span>
        </div>
        <div className="benefit">
          <span className="icon">💪</span>
          <span>Community-driven</span>
        </div>
      </div>
      <a href="https://bitcoin.org" target="_blank">Learn more about Bitcoin</a>
    </div>
  );
}
```

### 4. Tier Selection UI
**File:** `src/components/TierSelector.js` (new component)

```javascript
function TierSelector() {
  return (
    <div className="tier-selector">
      <h2>Choose Your WiFi Experience</h2>

      <div className="tiers">
        <div className="tier free">
          <h3>Free WiFi</h3>
          <p>Limited to 2Mbps</p>
          <p>Perfect for browsing & email</p>
          <button className="free-btn">Connect Free</button>
        </div>

        <div className="tier premium">
          <h3>Premium Unlimited</h3>
          <p>10 sats per hour</p>
          <p>Full speed WiFi access</p>
          <button className="premium-btn">Pay 10 sats</button>
        </div>
      </div>
    </div>
  );
}
```

### 5. Styling Updates
**File:** `src/styles/main.css`

```css
/* Trail's Coffee color scheme */
:root {
  --coffee-brown: #8B4513;
  --coffee-light: #D2B48C;
  --coffee-cream: #F5F5DC;
  --bitcoin-orange: #F7931A;
}

/* Header styling */
.header {
  background: linear-gradient(135deg, var(--coffee-brown), var(--coffee-light));
  color: white;
  text-align: center;
  padding: 2rem;
}

/* Payment options grid */
.payment-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  margin: 2rem 0;
}

.payment-option {
  border: 2px solid var(--coffee-brown);
  border-radius: 10px;
  padding: 1.5rem;
  text-align: center;
}

.payment-option h3 {
  color: var(--coffee-brown);
  margin-bottom: 1rem;
}

/* QR code styling */
.qr-code {
  width: 200px;
  height: 200px;
  margin: 1rem auto;
  border: 2px solid var(--bitcoin-orange);
  border-radius: 10px;
}
```

## Build Process

After making frontend changes:

```bash
# Navigate to frontend directory
cd portal-frontend/

# Install dependencies
npm install

# Build for production
npm run build

# Copy built files to the captive portal directory
cp -r build/* ../files/tollgate-captive-portal-site/
```

## Testing

1. Test QR code generation for both payment methods
2. Verify payment flows work correctly
3. Test responsive design on mobile devices
4. Verify branding displays correctly
5. Test educational content links

## Notes

- The backend already supports both Cashu and Lightning payments
- QR code generation libraries may need to be added to package.json
- Consider adding loading states and error handling for payment processing
