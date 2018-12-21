var app = new Vue({
    el: '#app',
    data: {
        time: '',
    },
    methods: {
        time: () => {
            axios
                .get('/api/time.go')
                .then(response => (this.time = response.data))
        }
    },
    mounted () {
        time
    }
});

