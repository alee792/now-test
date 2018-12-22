var clock = new Vue({
    el: '#clock',
    data: {
        time: '',
    },
    methods: {
        getTime() {
            axios
                .get('/api/time.go')
                .then(response => (this.time = response.data))
        },
        pollTime() {
            this.getTime()
            setTimeout(() => {
                this.pollTime();
            }, 1000)
        }
    },
    mounted () {
        this.pollTime()
    }
});

Vue.Component('user-profile', {
    props: ['user'],
    template: '<li>{{ user.login }}</li>'
})

var userApp = new Vue({
    el: '#userApp',
    data: {
        users: []
    },
    methods: {
        getUsers() {
            axios  
                .get('/api/users.go')
                .then(resp => (this.users = response.data))
        }
    },
    mounted () {

    }
});