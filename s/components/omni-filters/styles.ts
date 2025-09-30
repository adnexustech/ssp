import {css} from "@benev/slate"

export const styles = css`

:host {
	display: flex;
	height: 100%;
	overflow: scroll;
}

h2 {
	display: flex;

	& svg {
		width: 20px;
	}
}

.box {
	display: flex;
	align-items: flex-start;
	flex-direction: column;
	padding: 1em;

	& .dropdown {
		display: flex;
		flex-direction: column;

		& .flex {
			display: flex;
			align-items: center;
			margin: 1em 0;
		}
	}

	& label {
		font-size: 0.9em;
	}

	& select {
		background: #111;
		border-radius: 5px;
		color: gray;
		padding: 0.3em;

		& option {
			background: #111;
		}
	}
}

.filters {
	display: flex;
	flex-wrap: wrap;
	gap: 1em;

	&[disabled] {
		pointer-events: none;
		filter: blur(2px);
		opacity: 0.5;
	}

	& .filter {
		position: relative;

		& sl-dropdown, sl-button {
			width: 100%;
		}

		& sl-menu {
			width: 200px;
			padding: 0.5em;
		}
	}

	& .options {
		display: flex;
		flex-direction: column;
		padding: 0.5em;

		& fieldset {
			display: flex;
			flex-direction: column;
			padding: 0.3em;
		}
	}

	& button {
		font-family: "Inter";
		color: #fff;
		border: 1px solid #111;
		background-image: -webkit-gradient(
				linear,
				left bottom,
				left top,
				color-stop(0, rgb(48,48,48)),
				color-stop(1, rgb(102, 102, 102))
		);
		text-shadow: 0px -1px 0px rgba(0,0,0,.5);
		font-size: 0.8em;
		border-radius: 0;
		cursor: pointer;
		width: 100%;
	}

	& .filter-preview {
		position: relative;
		display: flex;
		flex-direction: column;
		width: 200px;
		height: 200px;
		justify-content: center;
		align-items: center;
		cursor: pointer;
		border: 1px solid #373535;
		border-radius: 5px;
		overflow: hidden;
		background: #1a1a1a;

		& .filter-name {
			position: absolute;
			bottom: 0;
			left: 0;
			right: 0;
			padding: 0.5em;
			background: rgba(0, 0, 0, 0.8);
			color: white;
			font-size: 0.85em;
			text-align: center;
			font-family: 'Inter', sans-serif;
			margin: 0;
		}

		& canvas {
			width: 100%;
			height: 100%;
			object-fit: cover;
		}

		&[data-selected] {
			border: 2px solid white;
			box-shadow: 0 0 10px rgba(255, 255, 255, 0.3);
		}

		&:hover {
			border-color: #666;
		}
	}

	& .filter-intensity {
		display: flex;
		flex-direction: column;
		padding: 0.5em;
		gap: 0.5em;

		& input {
			cursor: pointer;
		}
	}
}
`
