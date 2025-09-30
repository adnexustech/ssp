import {template, html, easypage} from "@benev/turtle"
export default template(async basic => {
	const path = basic.path(import.meta.url)

	return easypage({
		path,
		css: "index.css",
		title: "Adnexus Studio - Professional TV Ad Creation",
		head: html`
			<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.20.0/cdn/themes/dark.css" />
			<script type="module" src="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.20.0/cdn/shoelace.js" ></script>
			<!-- adding pixijs here because pixijs/filters work this way
				(because pixijs/filters does some global stuff work which
				doesnt seem to work when installing pixi through package json),
				but also adding pixijs to package json for types  -->
			<script src="https://cdn.jsdelivr.net/npm/pixi.js@7.4.2/dist/pixi.min.js"></script>
			<script src="https://cdn.jsdelivr.net/npm/pixi-filters@5.3.0/dist/browser/pixi-filters.min.js"></script>
			<script src="https://cdn.jsdelivr.net/npm/@pixi/graphics-extras@7.1.4/dist/graphics-extras.min.js"></script>
			<script src="https://cdn.jsdelivr.net/npm/@pixi-essentials/object-pool@1.0.1/dist/pixi-object-pool.js"></script>
			<script src="https://cdn.jsdelivr.net/npm/@pixi-essentials/bounds@3.0.0/dist/bounds.js"></script>
			<script src="https://cdn.jsdelivr.net/npm/@pixi-essentials/transformer@3.0.2/dist/transformer.js"></script>
			<!-- <script src="https://cdn.jsdelivr.net/npm/pixi.js@8.8.1/dist/pixi.min.js"></script> -->
			<!-- <script src="https://cdn.jsdelivr.net/npm/pixi-filters@6.1.0/dist/pixi-filters.min.js"></script> -->
			<script src="coi-serviceworker.js"></script>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<script type="importmap-shim" src="./importmap.json"></script>
			<script defer src="https://cdn.jsdelivr.net/npm/es-module-shims@1.8.2/dist/es-module-shims.min.js"></script>
			<script type="module-shim" src="./main.js"></script>
			<link rel="stylesheet" href="index.css">
			<link rel="icon" type="image/png" href="./assets/favicon.png">
			<link rel="preconnect" href="https://fonts.googleapis.com">
			<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
			<link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
			<meta name="description" content="Professional TV ad creation platform - Create premium 15s and 30s video ads with AI-powered tools">
			<meta property="og:title" content="Adnexus Studio - Professional TV Ad Creation">
			<meta property="og:description" content="Create broadcast-quality CTV/OTT ads in minutes with AI-powered video editing, QR codes, and call tracking">
			<meta property="og:type" content="website">
			<meta name="twitter:card" content="summary_large_image">
		`,
		body: html`
			<div class="loading-page-indicator">
				<img class="logo-loader" src="./assets/adnexus-logo-white.png" alt="Adnexus Studio" />
				<div class="loader"><div class="loaderBar"></div></div>
			</div>
		`
	})
})
