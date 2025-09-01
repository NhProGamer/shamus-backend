class GameWebSocket {
    constructor(onMessage) {
        this.socket = null;
        this.onMessageCallback = onMessage;
        this.connect();
    }

    connect() {
        this.socket = new WebSocket(`ws://${window.location.host}/game/ws`);

        this.socket.onopen = () => {
            console.log('WebSocket connecté');
            this.socket.send('{\n' +
                '    "channel": "game",\n' +
                '    "type": "gameData",\n' +
                '    "data": ""\n' +
                '}'
            );
        };

        this.socket.onmessage = (event) => {
            console.log(event);
            try {
                const data = JSON.parse(event.data);
                this.onMessageCallback(data);
            } catch (error) {
                console.error('Erreur de parsing du message', error);
            }
        };

        this.socket.onclose = (event) => {
            console.log('WebSocket déconnecté', event);
            //document.getElementById('connectionStatus').textContent = 'Déconnecté';
            setTimeout(() => this.connect(), 3000);
        };

        this.socket.onerror = (error) => {
            console.error('Erreur WebSocket', error);
            //document.getElementById('connectionStatus').textContent = 'Erreur';
        };
    }

    send(message) {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(message));
        } else {
            console.error('WebSocket non connecté');
        }
    }

    close() {
        if (this.socket) {
            this.socket.close();
        }
    }
}

let settings = {
    max_players: 5,
    roles: []
};
document.addEventListener('DOMContentLoaded', () => {
    // Slider range display
    const playerSlider = document.getElementById('playerSlider');
    const playerCount = document.getElementById('playerCount');
    playerSlider.value = settings.max_players;
    playerCount.textContent = settings.max_players

    playerSlider.addEventListener('input', () => {
        playerCount.textContent = playerSlider.value;
        settings.max_players = parseInt(playerSlider.value, 10);
    });

    // Role counter logic
    cards = document.querySelectorAll('.role-card')
    for (let i = 0; i < cards.length; i++) {
        const minusButton = cards[i].querySelector('.btn-minus');
        const plusButton = cards[i].querySelector('.btn-plus');
        const countSpan = cards[i].querySelector('.role-count');

        minusButton.addEventListener('click', () => {
            const currentValue = parseInt(countSpan.textContent, 10);
            if (currentValue > 0) {
                countSpan.textContent = currentValue - 1;
                settings.roles[i] = currentValue - 1;
            }
        });

        plusButton.addEventListener('click', () => {
            const currentValue = parseInt(countSpan.textContent, 10);
            countSpan.textContent = currentValue + 1;
            settings.roles[i] = currentValue + 1;
        });
    }
});

const gameSocket = new GameWebSocket((message) => {
    switch (message.channel) {
        case 'game':
            switch (message.type) {
                case 'gameData':
                    if (message.data.am_i_host) {
                        hostUI()
                    } else {
                        gameUI()
                    }
                    loadSettings({
                        max_players: message.data.max_players,
                        roles: message.data.roles
                    })
            }
            break;
    }
})

function hostUI() {
    document.getElementById('configMenu').style.display = '';
    document.getElementById('gameMenu').style.display = 'none';
}

function gameUI() {
    document.getElementById('configMenu').style.display = 'none';
    document.getElementById('gameMenu').style.display = '';
}

function sendSetttings() {
}

function loadSettings(set) {
    const playerSlider = document.getElementById('playerSlider');
    const playerCount = document.getElementById('playerCount');

    settings.max_players = set.max_players;
    settings.roles = set.roles;
    for (let i = 0; i < set.roles.length; i++) {
        cards[i].querySelector('.role-count').textContent = set.roles[i];
    }
    playerCount.textContent = set.max_players
    playerSlider.value = set.max_players
}