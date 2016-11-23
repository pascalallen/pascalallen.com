@extends('layouts.master')

@section('top-script')

	<style type="text/css">
		.row {
			/*text-align: center;*/
		}
		.user-image {
			width: 50px;
			border-radius: 20%;
			/*margin-left: auto;*/
   			/*margin-right: auto;*/
		}
	</style>

@stop

@section('content')

	<div class="row">
		<div class="col-md-8 col-md-offset-2">
			@foreach ($posts as $post)
				<table class="table">
					<thead>
						<tr>
							<th>Profile Image</th>
							<th>Post</th>
							<th>Date Posted</th>
							<th>User</th>
						</tr>
					</thead>
					<tbody>
						<tr>
							<td><a href="{{{ action('UsersController@show', $post->user->id)}}}"><img src="{{{ $post->user->image }}}" class="user-image"></a></td>	
							<td><a href="{{{ action('PostsController@show', $post->id) }}}">{{{ $post->title }}}</a></td>
							<td>{{{ $post->created_at->diffForHumans() }}}</td>
							<td><a href="{{{ action('UsersController@show', $post->user->id)}}}">{{{ $post->user->username}}}</a></td>
						</tr>
					</tbody>
				</table>
			@endforeach
			{{ $posts->links() }}
		</div>
	</div>

@stop

@section('bottom-script')

@stop