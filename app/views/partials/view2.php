
<div>

	Field:
	<br>
	<input type="text" data-ng-model="filter.name">

	<ul data-ng-controller="SimpleController">
		<li data-ng-repeat="friend in friends | filter:filter.name | orderBy:'city'"><% friend.name + " is from " + friend.city %></li>
	</ul>

</div>