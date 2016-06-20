<!DOCTYPE html>
<html lang="en">

<head>

    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="description" content="">
    <meta name="author" content="">

    <title>Drew Tattoo</title>

    <link rel="shortcut icon" href="img/Tattoo-Machine.ico">

    <!-- Bootstrap Core CSS -->
    <link href="css/bootstrap.min.css" rel="stylesheet">

    <!-- Custom CSS -->
    <link href="css/grayscale.css" rel="stylesheet">

    <!-- Custom Fonts -->
    <link href="font-awesome/css/font-awesome.min.css" rel="stylesheet" type="text/css">
    <link href="http://fonts.googleapis.com/css?family=Lora:400,700,400italic,700italic" rel="stylesheet" type="text/css">
    <link href="http://fonts.googleapis.com/css?family=Montserrat:400,700" rel="stylesheet" type="text/css">

    <!-- HTML5 Shim and Respond.js IE8 support of HTML5 elements and media queries -->
    <!-- WARNING: Respond.js doesn't work if you view the page via file:// -->
    <!--[if lt IE 9]>
        <script src="https://oss.maxcdn.com/libs/html5shiv/3.7.0/html5shiv.js"></script>
        <script src="https://oss.maxcdn.com/libs/respond.js/1.4.2/respond.min.js"></script>
    <![endif]-->

    @yield('top-script')

</head>

<body id="page-top" data-spy="scroll" data-target=".navbar-fixed-top">

    <!-- Navigation -->
    <nav class="navbar navbar-custom navbar-fixed-top" role="navigation">
        <div class="container">
            <div class="navbar-header">
                <button type="button" class="navbar-toggle" data-toggle="collapse" data-target=".navbar-main-collapse">
                    <i class="fa fa-bars"></i>
                </button>
                <a class="navbar-brand page-scroll" href="{{ action('HomeController@showWelcome') }}">
                    <i class="fa fa-anchor"></i>  <span class="light">Drew</span> Tattoo
                </a>
            </div>

            @if(Request::url() === 'http://blog.dev/tattoo-artist-project') <!-- Change to 'http://pascalallen.com/tattoo-artist-project' -->
                <!-- Collect the nav links, forms, and other content for toggling -->
                <div class="collapse navbar-collapse navbar-right navbar-main-collapse">
                    <ul class="nav navbar-nav">
                        <!-- Hidden li included to remove active class from about link when scrolled up past about section -->
                        <li class="hidden">
                            <a href="#page-top"></a>
                        </li>
                        <li>
                            <a class="page-scroll" href="#about">About</a>
                        </li>
                        <li>
                            <a class="page-scroll" href="{{ action('HomeController@showPortfolio') }}">Portfolio</a>
                        </li>
                        <li>
                            <a class="page-scroll" href="#contact">Contact</a>
                        </li>
                    </ul>
                </div>
                <!-- /.navbar-collapse -->
            @endif
        </div>
        <!-- /.container -->
    </nav>

    <!-- Intro Header -->
    <header class="intro">
        <div class="intro-body">
            <div class="container">
                <div class="row">
                    <div class="col-md-8 col-md-offset-2">
                        <h1 class="brand-heading">Drew Tattoo</h1>
                        <p class="intro-text">Schedule an appointment today!<br>Get some ink in your life.</p>
                        <a href="#about" class="btn btn-circle page-scroll">
                            <i class="fa fa-angle-double-down animated"></i>
                        </a>
                    </div>
                </div>
            </div>
        </div>
    </header>

    <!-- About Section -->
    <section id="about" class="container content-section text-center">
        <div class="row">
            <div class="col-lg-8 col-lg-offset-2">
                <h2>Drew Delafuente</h2>
                <p>Walking into 34-year-old Drew De La Fuente’s private studio at 1231 Austin Highway, clients instantly recognize that this local artist isn’t here to fool around. De La Fuente is known for his color and Japanese-style tattoos and has a reputation as one of the most precise tattooers in the area.</p>
                <p>He has been invited to prestigious tattoo conventions throughout the country and frequently sits as a guest artist in studios in Los Angeles and Chicago. A poster features De La Fuente’s work at a convention, sponsored by big-brand Eternal tattoo ink.</p>
                <p>De La Fuente’s artwork is big, bold and speaks volumes about the artist’s talent, yet this entrepreneur remains humble, always respecting his mentorship process and the artists who  brought him up in the industry in which he has worked for 16 years.</p>
                <h1 class="banner">Featured in the    <img src="img/express-news.png" class="express-news"></h1>
            </div>
        </div>
    </section>





    <section id="download" class="content-section text-center">
        <div class="download-section">
            <div class="container">
                <div class="col-lg-8 col-lg-offset-2">
                    <h2>Portfolio</h2>
                    <p>Check out some of my work and let's get you in for an appointment.</p>
                    <a href="{{ action('HomeController@showPortfolio') }}" class="btn btn-default btn-lg">Visit Portfolio Page</a>
                </div>
            </div>
        </div>
    </section>



    <!-- Contact Section -->
    <section id="contact" class="container content-section text-center">
        <div class="row">
            <div class="col-lg-8 col-lg-offset-2">
                <h2>Contact Drew</h2>
                <p>Call or email to schedule an appointment for an amazing tattoo today!</p>
                <p>(210) 763-0444</p>
                <p><a href="mailto:andrewdelaf@icloud.com">andrewdelaf@icloud.com</a>
                </p>
                <ul class="list-inline banner-social-buttons">
                    <li>
                        <a href="https://twitter.com/SBootstrap" class="btn btn-default btn-lg"><i class="fa fa-twitter fa-fw"></i> <span class="network-name">Twitter</span></a>
                    </li>
                    <li>
                        <a href="https://www.instagram.com/drewtattoo13/" class="btn btn-default btn-lg"><i class="fa fa-instagram fa-fw"></i> <span class="network-name">Instagram</span></a>
                    </li>
                    <li>
                        <a href="https://www.facebook.com/drew.delafuente1" class="btn btn-default btn-lg"><i class="fa fa-facebook-square fa-fw"></i> <span class="network-name">Facebook</span></a>
                    </li>
                </ul>
            </div>
        </div>
    </section>

    <!-- Map Section -->
    <div id="map"></div>
    
    <!-- Footer -->
    <footer>
        <div class="container text-center">
            <p>Copyright &copy; Drew Delafuente 2016</p>
        </div>
    </footer>
    
    <!-- jQuery -->
    <script src="js/jquery.js"></script>

    <!-- Bootstrap Core JavaScript -->
    <script src="js/bootstrap.min.js"></script>

    <!-- Plugin JavaScript -->
    <script src="js/jquery.easing.min.js"></script>

    <!-- Google Maps API Key - Use your own API key to enable the map feature. More information on the Google Maps API can be found at https://developers.google.com/maps/ -->
    <script type="text/javascript" src="https://maps.googleapis.com/maps/api/js?key=AIzaSyBDNUiHgyFcQO7Oe9Li4hzEwh_6hTAp3yM&sensor=false"></script>

    <!-- Custom Theme JavaScript -->
    <script src="js/grayscale.js"></script>

    @yield('bottom-script')

</body>

</html>
