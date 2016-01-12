<?php

class BaseController extends Controller {

	public function __construct()
	{
	    // require csrf token for all post, delete, and put actions
	    $this->beforeFilter('csrf', array('on' => array('post', 'delete', 'put')));
	}

	/**
	 * Setup the layout used by the controller.
	 *
	 * @return void
	 */
	protected function setupLayout()
	{
		if ( ! is_null($this->layout))
		{
			$this->layout = View::make($this->layout);
		}
	}

}
