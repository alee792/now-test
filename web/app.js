Vue.component('user-profile', {
    props: ['user'],
    template: '<li> {{ user.Login }} {{ user.Total }} </li>'
})

var clock = new Vue({
    el: '#clock',
    data: {
        time: '',
    },
    methods: {
        getTime() {
            axios
                .get('/api/time.go')
                .then(resp => (this.time = resp.data))
        },
        pollTime() {
            this.getTime()
            setTimeout(() => {
                this.pollTime();
            }, 1000)
        }
    },
    mounted () {
        this.getTime()
    }
});

var userApp = new Vue({
    el: '#user-app',
    data: {
        users: []
    },
    methods: {
        getUsers() {
            axios  
                .post('/api/user.go',{
                    owner: "go-chi",
                    repo: "chi"
                })
                .then(resp => (this.users = resp.data))
        }
    },
    computed: {
        sortedUsers: function() {
            function compare(a,b) {
                if (a.Total < b.Total)
                    return 1;
                if (a.Total > b.Total)
                    return -1;
                return 0;
            }
            return this.users.sort(compare);
        }
    },
    mounted () {
        this.getUsers()
    }
});