@extends('layouts.master')

@section('content')

<div class="container">
	<h2>Blog Posts</h2>
	<p>Check it out!</p>            
	<table class="table table-condensed">
		<thead>
			<tr>
				<th>Title</th>
				<th>Body</th>
			</tr>
		</thead>
		<tbody>
				@foreach ($posts as $post)
					<tr>
						<td>{{{ $post->title }}}</td>
						<td>{{{ $post->body }}}</td>
					</tr>
				@endforeach
		</tbody>
	</table>
</div>

@stop