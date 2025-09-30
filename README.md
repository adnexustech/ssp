<p align="center"><img width="300" src="./assets/icon2.png"/></p>
<p align="center"><a href="https://opensource.org/license/mit"><img src="https://img.shields.io/badge/license-MIT-blue.svg"/></a></p>
<p align="center">Professional TV Ad Creation Platform</p>

# Adnexus Studio

Create premium 15-second and 30-second TV ads in minutes. No production crews, no agencies, no complexity.

## Introduction

Adnexus Studio is a professional video ad creation platform that runs entirely in your browser. Built on cutting-edge web technologies, it empowers advertisers to create broadcast-quality CTV/OTT ads with AI-powered tools and native integrations with Zen AI models.

**Key Features:**
- 🎬 Professional 15s & 30s TV ad templates
- 🤖 AI-powered video editing with Zen models
- 📱 QR code generation for direct response campaigns
- 📞 Twilio integration for call tracking
- 🚀 Seamless integration with Adnexus DSP
- 🔒 Privacy-first: All editing happens locally in your browser
- ⚡ High-performance WebCodecs rendering

## The Adnexus Ecosystem

Adnexus Studio is part of the complete Adnexus advertising technology stack:

- **Adnexus Studio** (this product) - Create your TV ads
- **Adnexus DSP** - Launch campaigns across 500+ channels
- **Adnexus SSP** - Monetize your CTV/OTT inventory
- **Adnexus Exchange** - High-performance programmatic ad marketplace

## Features

### Core Video Editing
- ✂️ Trimming & Splitting
- 📹 Multi-format support (MP4, MOV, and more)
- 🖼️ Image, text, audio, and video layers
- 🎨 Visual effects and transitions
- 📐 Clip editing (rotate, resize, styling)
- ⏪ Undo/Redo
- 🎯 Render up to 4K resolution
- 🎬 10-120 FPS timebase options
- 💾 Project manager for saving multiple projects

### TV Ad-Specific Features
- ⏱️ **15s & 30s Templates** - Pre-built templates optimized for TV advertising
- 📱 **QR Code Generator** - Add trackable QR codes for direct response
- 📞 **Twilio Integration** - Dynamic phone number insertion for call tracking
- 🤖 **AI Video Tools** - Powered by Zen AI for content generation and editing
- 🎯 **CTV Optimization** - Export formats optimized for Connected TV platforms
- 🏷️ **Brand Asset Management** - Store and reuse logos, colors, and fonts

### AI-Powered by Zen
Integrated with the Zen AI model family from `~/work/zen/`:
- 🖼️ AI image generation and editing
- 🎥 Video content generation
- ✍️ Automated copy generation
- 🎨 Style transfer and effects
- 🗣️ Multi-language support

### Advanced Capabilities
- 🎭 GL Transitions for smooth visual effects
- 🎛️ Advanced audio editing
- 🔄 Real-time collaboration via WebRTC
- 🎬 Keyframe animations (coming soon)
- 🗣️ Speech-to-text (coming soon)

## Quick Start

### Using Adnexus Studio Online
Visit [studio.ad.nexus](https://studio.ad.nexus) to start creating ads instantly.

### Development Setup

1. Clone the repository:
```sh
git clone git@github.com:adnexusinc/studio.git
cd studio
```

2. Install dependencies:
```sh
npm install
```

3. Build the project:
```sh
npm run build
```

4. Start development server:
```sh
npm start
```

Visit `http://localhost:8000` to access the studio.

## Using Studio Components

Embed Adnexus Studio components in your own applications:

### Installation
```sh
npm install @adnexus/studio
```

### Import & Register
```js
import {getComponents, registerElements} from '@adnexus/studio'
registerElements(getComponents())
```

### Use Components
```html
<adnexus-text></adnexus-text>
<adnexus-media></adnexus-media>
<adnexus-timeline></adnexus-timeline>
<adnexus-qrcode></adnexus-qrcode>
```

## Architecture

Adnexus Studio follows a unidirectional data flow architecture:

```
Actions → State → Controllers → Components/Views
```

### Key Components:
1. **State** - Centralized application state
2. **Actions** - State mutation functions
3. **Controllers** - Business logic layer
4. **Components** - LitElement-based UI components

## Tech Stack

- **Core**: TypeScript, LitElement
- **Video Processing**: FFmpeg.wasm, WebCodecs API
- **Rendering**: PixiJS, GSAP
- **AI Integration**: Zen models (Qwen3-based)
- **Communication**: Twilio SDK
- **QR Codes**: qrcode library
- **Build**: Rollup, Turtle

## Integration with Adnexus DSP

Once you create your ad in Studio:

1. **Export** your 15s or 30s video ad
2. **Upload** directly to Adnexus DSP
3. **Launch** campaigns across 500+ premium channels
4. **Track** performance with QR codes and call tracking

## Browser Requirements

⚠️ Adnexus Studio requires modern browsers with WebCodecs support:
- Chrome/Edge 94+
- Safari 16.4+
- Firefox (limited support)

Older browsers may not work properly due to reliance on cutting-edge web APIs.

## Contributing

We welcome contributions! To contribute:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

Join our development discussions: [Discord](https://discord.gg/adnexus)

### Development Guidelines
- Follow TypeScript best practices
- Write tests for new features
- Update documentation
- Follow the existing architecture patterns

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Contact

- **Website**: [ad.nexus](https://ad.nexus)
- **Studio**: [studio.ad.nexus](https://studio.ad.nexus)
- **Support**: support@ad.nexus
- **Sales**: sales@ad.nexus

---

<p align="center">
  <strong>Built with ❤️ by Adnexus Technology Inc</strong><br>
  Part of the Adnexus Ad Tech Ecosystem
</p>