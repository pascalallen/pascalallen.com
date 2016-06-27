<!doctype html>
<html>
<head>
	<title>My Calculator</title>
	<link href = "css/calculator.css" rel = "stylesheet">
</head>
<body>
	<div id = "calculator">
		<div class = "top">
			<span class = "clear" id = "clear">C</span>
			<div class = "screen" id = "screen1"></div>
			<div class = "screen2" id = "screen2"></div>
			<div class = "screen3" id = "screen3"></div>
		</div>
		<div class="keys">
			<span data-value = "7" class = "number">7</span>
			<span data-value = "8" class = "number">8</span>
			<span data-value = "9" class = "number">9</span>
			<span data-value = "+" class = "operator">+</span>
			<span data-value = "4" class = "number">4</span>
			<span data-value = "5" class = "number">5</span>
			<span data-value = "6" class = "number">6</span>
			<span data-value = "-" class = "operator">-</span>
			<span data-value = "1" class = "number">1</span>
			<span data-value = "2" class = "number">2</span>
			<span data-value = "3" class = "number">3</span>
			<span data-value = "/" class = "operator">/</span>
			<span data-value = "0" class = "number">0</span>
			<span data-value = "." class = "number">.</span>
			<span data-value = "=" class = "eval" id = "eval">=</span>
			<span data-value = "*" class = "operator">*</span>
		</div>
	</div>
<script src = "js/calculator.js"></script>
</body>
</html>