@extends('layouts.master')

@section('top-script')

		<style type="text/css">

			.row {
				text-align: center;
			}

		</style>

@stop

@section('content')

	<div class="row">
		<div class="col-md-8 col-md-offset-2">
			@if(Auth::user() != $user)
				{{{ $user->username }}}
				<img src="{{{ $user->image }}}">
			@else
				<h1>Let's do this, {{{ $user->username }}}!</h1>
				<br>
				<img src="{{{ $user->image }}}">
				<br>
				<br>
				<a href="{{{action('PostsController@create')}}}"><i class="fa fa-pencil fa-3x"></i></a>
			@endif
			<br>
			<h1>{{{ $user->username }}}'s Content</h1>

					@foreach($user->post()->get() as $post)
						<table class="table">
							<tbody>
								<tr>
									<td><a href="{{{ action('PostsController@show', $post->id) }}}">{{{ $post->title }}}</a></td>
									@if(Auth::id() == $user->id)
										<td><a href="{{{ action('PostsController@edit', $post->id) }}}">{{ Form::submit('Edit', ['class' => 'btn btn-warning']) }}</a></td>
										{{ Form::open(array('action' => array('PostsController@destroy', $post->id, 'files' => true), 'method' => 'DELETE')) }}
											<td>{{ Form::submit('Delete', ['class' => 'btn btn-danger']) }}</td>
										{{Form::close()}}
									@endif
								</tr>
							</tbody>
						</table>
					@endforeach
		</div>
	</div>

@stop

@section('bottom-script')

@stop

