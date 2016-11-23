"use strict";
$(document).ready(function() {
	$(".nav a").on("click", function(){
		$(".nav").find(".active").removeClass("active");
		$(this).parent().addClass("active");
	});


	var myIndex = 0;
	carousel();

	function carousel() {
	    var i;
	    var x = document.getElementsByClassName("mySlides");
	    for (i = 0; i < x.length; i++) {
	       x[i].style.display = "none";
	    }
	    myIndex++;
	    if (myIndex > x.length) {myIndex = 1}
	    x[myIndex-1].style.display = "block";
	    setTimeout(carousel, 300); // Change image every 2 seconds
	}

	

});


var sampleApp = angular.module('sampleApp', [], function($interpolateProvider) {
	$interpolateProvider.startSymbol('<%');
    $interpolateProvider.endSymbol('%>');
});

sampleApp.config(function($routeProvider){
	$routeProvider
		.when('/view1', 
			{
				templateUrl: 'partials/view1.php',
				controller: 'SimpleController'
			})
		.when('/view2', 
			{
				templateUrl: 'partials/view2.php',
				controller: 'SimpleController'
			})
		.otherwise({ redirectTo: '/view1' });
});

// var sampleControllers = {};

// sampleControllers.SimpleController = function($scope){
// 	$scope.friends = [
//     	{name:'Dave', city:'Cleveland'}, 
//     	{name:'Pascal', city:'Demopolis'}, 
//     	{name:'Sean', city:'Tokyo'}, 
//     	{name:'Coleman', city:'Kansas City'}
// 	];
// };

// sampleControllers.addFriend = function($scope){
// 	$scope.friends.push({ 
// 		name: $scope.newFriend.name, 
// 		city: $scope.newFriend.city 
// 	});
// }

// sampleApp.controller(sampleControllers);

sampleApp.controller('SimpleController', function($scope){
	$scope.friends = [
    	{name:'Dave', city:'Cleveland'}, 
    	{name:'Pascal', city:'Demopolis'}, 
    	{name:'Sean', city:'Tokyo'}, 
    	{name:'Coleman', city:'Kansas City'}
	];

	$scope.addFriend = function(){
		$scope.friends.push({ 
			name: $scope.newFriend.name, 
			city: $scope.newFriend.city 
		});
	};
});




