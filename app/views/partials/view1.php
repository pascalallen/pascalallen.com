
<div class="container">

	Field:
	<br>
	<input type="text" data-ng-model="filter.name">

	<ul data-ng-controller="SimpleController">
		<li data-ng-repeat="friend in friends | filter:filter.name | orderBy:'city'"><% friend.name + " is from " + friend.city %></li>
	</ul>
	<br>
	Friend name:
	<br>
	<input type="text" data-ng-model="newFriend.name">
	<br>
	Friend city:
	<br>
	<input type="text" data-ng-model="newFriend.city">
	<br>
	<button data-ng-click="addFriend()">Add Friend</button>
	<br>
	<a href="#/view2">View 2</a>
</div>