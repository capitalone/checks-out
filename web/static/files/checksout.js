(function () {

	/**
	 * Creates the angular application.
	 */
	var app = angular.module('app', [
			'ngRoute',
			'ngDialog',
			'toggle-switch'
		]);

	app.config(["ngDialogProvider", function (ngDialogProvider) {
		ngDialogProvider.setDefaults({
			className: "ngdialog-theme-default",
			plain: false,
			showClose: true,
			closeByDocument: true,
			closeByEscape: true,
			appendTo: false,
			preCloseCallback: function () {
				console.log("default pre-close callback");
			}
		});
	}]);

	/**
	 * Defines the route configuration for the
	 * main application.
	 */
	function Config ($routeProvider, $httpProvider, $locationProvider) {
		$routeProvider
		.when('/', {
			templateUrl: '/static/checksout.html',
			controller: 'RepoCtrl'
		})
		.when('/:org', {
			templateUrl: '/static/checksout.html',
			controller: 'RepoCtrl'
		})
		;

		// Enables html5 mode
		$locationProvider.html5Mode(true);

		// Enables XSRF protection
		$httpProvider.defaults.headers.common['X-CSRF-TOKEN'] = window.STATE_FROM_SERVER.csrf;
	}

	function Noop($rootScope) {}

	angular
		.module('app')
		.config(Config)
		.run(Noop);
})();

(function () {

	function parseRepo() {
		return function(conf_url) {
			var parts = conf_url.split("/");
			return parts[3]+"/"+parts[4];
		}
	}

	angular
		.module('app')
		.filter('parseRepo', parseRepo);

})();

(function () {
	function PropService($http) {
	    var docsUrl_ = window.STATE_FROM_SERVER.docsUrl;

	    this.docsUrl = function() {
	        return docsUrl_;
	    }
	}

	angular
		.module('app')
		.service('prop', PropService);
})();

(function () {
	function UserService($http) {
		var user_ = window.STATE_FROM_SERVER.user;
		var deleted_ = false;

		this.current = function() {
			return user_;
		};

		this.deleted = function() {
			return deleted_;
		};

		this.delete = function() {
			_deleted = true;
			return $http.delete('/api/user');
		};
	}

	angular
		.module('app')
		.service('user', UserService);
})();

(function () {
	function TeamService($http) {
		var teams_ = window.STATE_FROM_SERVER.teams || [];
		teams_.unshift(window.STATE_FROM_SERVER.user);

		this.list = function() {
			return teams_;
		};

		this.get = function(name) {
			for (var i=0; i<teams_.length; i++) {
				if (teams_[i].login === name) {
					return teams_[i];
				}
			}
		}
	}

	angular
		.module('app')
		.service('teams', TeamService);
})();

(function () {
	function RepoService($http) {

		this.list = function(userId, orgId) {
			if (userId === orgId) {
				return $http.get('/api/user/repos');
			} else {
				return $http.get('/api/user/repos/' + orgId);
			}
		};

		this.post = function(repo, body) {
			return $http.post('/api/repos/'+repo.owner+'/'+repo.name, body);
		};

		this.delete = function(repo) {
			return $http.delete('/api/repos/'+repo.owner+'/'+repo.name);
		};

		this.validate = function(repo) {
			return $http.get('/api/repos/'+repo.owner+'/'+repo.name+'/validate');
		};
	}

	angular
		.module('app')
		.service('repos', RepoService);
})();

(function () {
    function OrgService($http) {

        this.list = function() {
            return $http.get('/api/user/orgs/enabled');
        };

        this.delete = function(org) {
        	return $http.delete('/api/repos/'+org.login);
		};

        this.add = function(org) {
        	return $http.post('/api/repos/'+org.login, {});
		};
	}

    angular
        .module('app')
        .service('orgs', OrgService);
})();

(function () {
	function RepoCtrl($scope, $routeParams, $location, repos, teams, user, prop, orgs, ngDialog) {

		$scope.refresh = function() {
            repos.list($scope.user.login, $scope.org.login).then(function(payload){
                $scope.repos = payload.data;
                delete $scope.error;
            }).catch(function(err){
                $scope.error = err;
            });

            orgs.list().then(function(payload){
                //not efficient, but probably not too many enabled orgs
                //can be revisited if slow
                for (var i = 0;i< payload.data.length;i++) {
                    var item = payload.data[i];
                    for (var j = 0;j<$scope.orgs.length;j++) {
                        if ($scope.orgs[j].login === item.login) {
                            $scope.orgs[j].enabled = true;
                            break;
                        }
                    }
                }
            }).catch(function(err) {
                $scope.error = err;
            })
		};

		if (!user.deleted()) {

			$scope.org = teams.get($routeParams.org || user.current().login);
			$scope.orgs = teams.list();
			$scope.user = user.current();
			$scope.docsUrl = prop.docsUrl();

			$scope.refresh();
		}

		$scope.activate = function(repo) {
			var index = $scope.repos.indexOf(repo);
			$scope.saving = true;
			repos.post(repo, {}).then(function(payload){
				delete $scope.repo;
				delete $scope.error;
				$scope.repos[index] = payload.data;
				$scope.saving = false;
			}).catch(function(err){
				delete $scope.repo;
				$scope.error = err;
				delete repo.id;
				$scope.saving = false;
			});
		};

        $scope.activateOrg = function(org) {
            $scope.saving = true;
            orgs.add(org).then(function(payload){
                delete $scope.repo;
                delete $scope.error;
                $scope.saving = false;
                $scope.refresh();
            }).catch(function(err){
                delete $scope.repo;
                $scope.error = err;
                org.enabled = false;
                $scope.saving = false;
            });
        };

		$scope.delete = function(repo) {
			delete repo.id;
			repos.delete(repo).catch(function(err){
				$scope.error = err;
			});
		};

        $scope.deleteOrg = function(org) {
            org.enabled = false;
            orgs.delete(org).then(function(payload){
                delete $scope.repo;
                delete $scope.error;
                $scope.saving = false;
                $scope.refresh();
            }).catch(function(err){
                $scope.error = err;
                org.enabled= true;
            });
        };

		$scope.deleteUser = function() {
			user.delete().then(function(payload){
				window.location.href = '/logout';
			}).catch(function(err) {
				$scope.error = err;
			});
		};

		$scope.changeOrg = function(value) {
			$scope.org = teams.get(value);
		};

		$scope.edit = function(repo) {
			$scope.repo = repo;
		};

		$scope.toggle = function(repo) {
			if (!repo.id) {
				$scope.delete(repo);
			} else {
				ngDialog.openConfirm({
					template:'/_confirm_template',
					className: 'ngdialog-theme-default',
					scope: $scope,
				}).then(function (value) {
					$scope.activate(repo);
				}, function (reason) {
					delete repo.id;
				});
			}
		};

        $scope.toggleOrg = function(org) {
            if (!org.enabled) {
                $scope.deleteOrg(org);
            } else {
				ngDialog.openConfirm({
					template:'/_confirm_template',
					className: 'ngdialog-theme-default',
					scope: $scope,
				}).then(function (value) {
					$scope.activateOrg(org);
				}, function (reason) {
					delete $scope.repo;
					org.enabled = false;
				});
            }
        };

		$scope.close = function() {
			delete $scope.repo;
			delete $scope.validInfo;
		};

		$scope.validate = function(repo) {
			repos.validate(repo).then(function(payload){
				payload.slug = repo.slug;
				$scope.validInfo=payload;
				var d = $scope.validInfo.data;
				//d should be a json with two fields, message and file. file will be empty if there's no file to convert
				//message should always be populated
                $scope.validInfo.message = d.message;
				if (d.file.length > 0) {
				    $scope.validInfo.fileContent = d.file;
                }
			}).catch(function(err) {
				err.slug = repo.slug;
				$scope.validInfo = {
				    slug: repo.slug,
				    message: err.data
				};
			});
		};

		$scope.saving = false;
	}

	angular
		.module('app')
		.controller('RepoCtrl', RepoCtrl);
})();
