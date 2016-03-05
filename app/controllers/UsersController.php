<?php

class UsersController extends \BaseController {

	public function __construct()
	{
		parent::__construct();

		$this->beforeFilter('auth', array('except' => array('create', 'store')));
	}
	/**
	 * Display a listing of the resource.
	 *
	 * @return Response
	 */
	public function index()
	{
		$user = Auth::user();
		return View::make('users.index')->with('user', $user);
	}


	/**
	 * Show the form for creating a new resource.
	 *
	 * @return Response
	 */
	public function create()
	{
		return View::make('users.create');
	}


	/**
	 * Store a newly created resource in storage.
	 *
	 * @return Response
	 */
	public function store()
	{
		$user = new User();
        Log::info(Input::all());

		$validator = Validator::make(Input::all(), User::$rules);

		if ($validator->fails()) {
	        // validation failed, redirect to the tutorial create page with validation errors and old inputs
	        Session::flash('errorMessage', 'Validation failed');
	        return Redirect::back()->withInput()->withErrors($validator);
	    } else {
			$user->username = Input::get('username');
			$user->password = Input::get('password');
			$user->email = Input::get('email');
			$image = Input::file('image');
			$user->location = Input::get('location');
			
			if ($image) {
				$filename = $image->getClientOriginalName();
				$user->image = '/uploaded/' . $filename;
				$image->move('uploaded/', $filename);
			} else{
				$user->image="/css/monkey-icon-taupe-on-cream.png";
			}

			$user->save();

			Auth::login($user);
			$user = Auth::user();

			Session::flash('successMessage', 'Your user has been saved.');
			return Redirect::action('UsersController@show', $user->id);
		}
	}


	/**
	 * Display the specified resource.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function show($id)
	{
		$user = User::find($id);
		
		return View::make('users.show', compact('user'));
	}


	/**
	 * Show the form for editing the specified resource.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function edit($id)
	{
		if(Auth::id() != $id)
		{
			Log::info('This user id is not equal');
			return Redirect::action('UsersController@show', Auth::id());
		}
		$post = Post::with('user')->get();
		$user = User::find($id);
		return View::make('users.edit', compact('user'));
	}


	/**
	 * Update the specified resource in storage.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function update($id)
	{
		$user = User::find($id);
		Log::info(Input::all());

		$validator = Validator::make(Input::all(), User::$rules);

		if ($validator->fails()) {
	        // validation failed, redirect to the tutorial create page with validation errors and old inputs
	        Session::flash('errorMessage', 'Validation failed');
	        return Redirect::back()->withInput()->withErrors($validator);
	    } else {
			$user->username = Input::get('username');
			$user->password = Input::get('password');
			$user->email = Input::get('email');

			$image = Input::file('image');
			if ($image) {
				$filename = $image->getClientOriginalName();
				$user->image = '/uploaded/' . $filename;
				$image->move('uploaded/', $filename);
			}

			$user->save();

			Auth::login($user);
			$user = Auth::user();

			Session::flash('successMessage', 'Your user has been saved.');
			return Redirect::action('UsersController@show', $user->id);
		}
	}


	/**
	 * Remove the specified resource from storage.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function destroy($id)
	{
		$user = User::find($id);
		$user->delete();

		return Redirect::action('HomeController@showWelcome');
	}


}
