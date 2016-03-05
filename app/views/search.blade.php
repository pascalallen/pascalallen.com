@extends('layouts.master')

@section('top-script')

	<style type="text/css">
		.row {
			text-align: center;
		}

		.user-image {
			width: 50px;
			height: 50px;
			/*margin-left: auto;*/
   			/*margin-right: auto;*/
		}
	</style>

@stop

@section('content')

	<div class="row">
		<div class="col-md-8 col-md-offset-2">
			@foreach ($resultsPost as $result)
			<h1>Posts</h1>
				<table class="table">
					<thead>
						<tr>
							<th>Profile Image</th>
							<th>Post</th>
							<th>Date Posted</th>
							<th>User</th>
							<th>Location</th>
						</tr>
					</thead>
					<tbody>
						<tr>
							<td><a href="{{{ action('UsersController@show', $result->user->id) }}}"><img src="{{{ $result->user->image }}}" class="user-image"></a></td>	
							<td><a href="{{{ action('HomeController@searchShow', $result->id) }}}">{{{ $result->title }}}</a></td>
							<td>{{{ $result->created_at->diffForHumans() }}}</td>
							<td><a href="{{{ action('UsersController@show', $result->user->id) }}}">{{{ $result->user->username }}}</a></td>
							<td>{{{ $result->user->location }}}</a></td>
						</tr>
					</tbody>
				</table>
			@endforeach
		</div>
	</div>
	<div class="row">
		<div class="col-md-4 col-md-offset-4">
			@foreach ($resultsUser as $result)
			<h1>Users</h1>
				<table class="table">
					<thead>
						<tr>
							<th>Profile Image</th>
							<th>Last Post</th>
							<th>User</th>
							{{-- <th>Location</th> --}}
						</tr>
					</thead>
					<tbody>
						<tr>
							<td><a href="{{{ action('UsersController@show', $result->id) }}}"><img src="{{{ $result->image }}}" class="user-image"></a></td>	
							<td>{{{ $result->created_at->diffForHumans() }}}</td>
							<td><a href="{{{ action('UsersController@show', $result->id) }}}">{{{ $result->username }}}</a></td>
							{{-- <td>{{{ $result->location }}}</a></td> --}}
						</tr>
					</tbody>
				</table>
			@endforeach
		</div>
	</div>

@stop

@section('bottom-script')

@stop