<form role="form" method="POST">
  <div class="form-group">
    <label for="email">Email address:</label>
    <input type="email" class="form-control" id="email" name="email" value="{{{ Input::old('email') }}}">
  </div>
  <div class="form-group">
    <label for="pwd">Password:</label>
    <input type="password" class="form-control" id="pwd" name="password" value="{{{ Input::old('password') }}}">
  </div>
  <div class="checkbox">
    <label><input type="checkbox"> Remember me</label>
  </div>
  <button type="submit" class="btn btn-default">Submit</button>
</form>



{{-- if (all is good) {
  	continue
} else {
	return Redirect::back()->withInput();

	// or:

	return Redirect::action('PostsController@create')->withInput();
} --}}