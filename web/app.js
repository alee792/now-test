var app = new Vue({
    el: '#app',
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

