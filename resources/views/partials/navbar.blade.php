<nav class="navbar navbar-default">
	<div class="container-fluid">
		<!-- Brand and toggle get grouped for better mobile display -->
		<div class="navbar-header">
			<button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#bs-example-navbar-collapse-1" aria-expanded="false">
				<span class="sr-only">Toggle navigation</span>
				<span class="icon-bar"></span>
				<span class="icon-bar"></span>
				<span class="icon-bar"></span>
			</button>
			<a class="navbar-brand" href="{!! action('HomeController@showWelcome') !!}">User - Admin</a>
		</div>

		<!-- Collect the nav links, forms, and other content for toggling -->
		<div class="collapse navbar-collapse" id="bs-example-navbar-collapse-1">
			<ul class="nav navbar-nav">
				@if(Auth::check())
					<li><a href="">New Post</a></li>
					<li class="dropdown">
						<a href="#" class="dropdown-toggle" data-toggle="dropdown" role="button" aria-haspopup="true" aria-expanded="false">Account <span class="caret"></span></a>
						<ul class="dropdown-menu">
							<li><a href="{!! action('UsersController@show', Auth::user()->id) !!}">Profile</a></li>
							<li><a href="{!!action('UsersController@edit', Auth::id())!!}">Account Settings</a></li>
							<li role="separator" class="divider"></li>
							<li><a href="{!! action('HomeController@getLogout') !!}">Log Out</a></li>
						</ul>
					</li>
				@endif
			</ul>
			{!! Form::open(array('action' => array('HomeController@search'), 'method' => 'GET', 'class' => 'navbar-form navbar-left', 'role' => 'search')) !!}
				<div class="form-group">
					{!! Form::text('search', null, ['class' => 'form-control', 'placeholder' => 'Search']) !!}
				</div>
				{!! Form::submit('submit', ['class' => 'btn btn-default']) !!}
			{!! Form::close() !!}
			<ul class="nav navbar-nav navbar-right">
				@if(!Auth::check())
					<li><a href="{!! action('HomeController@getLogin') !!}">Log In</a></li>
					<li><a href="{!! action('UsersController@create') !!}">Sign Up!</a></li>
				@endif
				<li><a href="">Posts</a></li>
			</ul>
		</div><!-- /.navbar-collapse -->
	</div><!-- /.container-fluid -->
</nav>