<?php

class HomeController extends BaseController {

	/*
	|--------------------------------------------------------------------------
	| Default Home Controller
	|--------------------------------------------------------------------------
	|
	| You may wish to use controllers instead of, or in addition to, Closure
	| based routes. That's great! Here is an example controller method to
	| get you started. To route to this controller, just add the route:
	|
	|	Route::get('/', 'HomeController@showWelcome');
	|
	*/

	public function showWelcome()
	{
		return Redirect::action('HomeController@sayHello', array('Bob'));
	}

	public function showResume()
	{
		return View::make('resume');
	}

	public function showPortfolio()
	{
		return View::make('portfolio');
	}

	public function showName($name = 'Pascal')
	{
		return $name;
	}

	public function sayHello($name = 'Pascal')
	{
		if ($name == "Chris") {
        	return Redirect::to('/');
    	} else {
    		$data = array('name' => $name);
        	return View::make('my-first-view')->with($data);
    	}
	}

	public function rollDice($guess = '1')
	{
		$randNum = rand(1, 6);
		$data = array(
			'guess' => $guess, 
			'randNum' => $randNum
		);
    	return View::make('roll-dice')->with($data);
	}
}
