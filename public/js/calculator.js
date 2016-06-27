(function() {
"use strict";
	var numbers = document.getElementsByClassName('number');
	var operator = document.getElementsByClassName('operator');
	var screen1 = document.getElementById('screen1');
	var screen2 = document.getElementById('screen2');
	var screen3 = document.getElementById('screen3');
	var activeScreen = screen1;

	var displayNumber = function () { 
		activeScreen.innerHTML = activeScreen.innerHTML + this.getAttribute('data-value');
	};

	var displayOp = function () {
		screen2.innerHTML = this.getAttribute('data-value');
		activeScreen = screen3;
	};

	var clearScreens = function () { 
		screen1.innerHTML = '';
		screen2.innerHTML = '';
		screen3.innerHTML = '';
		activeScreen = screen1;
	};

	var clear = document.getElementById('clear');
	clear.addEventListener('click', clearScreens);

	var displayEval = function () {
		var a = parseFloat(screen1.innerHTML);
		var b = parseFloat(screen3.innerHTML);
		var answer;
		switch (screen2.innerHTML) {
		    case "+":
		        answer = a + b;
		        break;
		    case "-":
		        answer = a - b;
		        break;
		    case "/":
		        answer = a / b;
		        break;
		    case "*":
		        answer = a * b;
		        break;
		};
		clearScreens();
		activeScreen.innerHTML = answer;
	};

	var equals = document.getElementById('eval');
	equals.addEventListener('click', displayEval);

	for (var i = 0; i < numbers.length; i++) {
		numbers[i].addEventListener('click', displayNumber);
	};

	for (var x = 0; x < operator.length; x++) {
		operator[x].addEventListener('click', displayOp);
	};

})();