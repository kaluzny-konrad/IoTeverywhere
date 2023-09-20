import random
import time
import os

# Zasymulowana klasa CPUTemperature
class MockCPUTemperature:
    def __init__(self):
        pass

    @property
    def temperature(self):
        # Generowanie losowej temperatury otoczenia od 10°C do 30°C
        return round(random.uniform(10.0, 30.0), 2)

def create_or_open_file(filename):
    if not os.path.isfile(filename):
        # Jeśli plik nie istnieje, to go tworzymy
        with open(filename, "w") as log:
            log.write("Timestamp,Temperature\n")
    return filename

try:
    while True:
        mock_cpu = MockCPUTemperature()
        temp = mock_cpu.temperature

        # Pobierz aktualną datę i godzinę w formacie "YYYYMMDD_HHMMSS"
        current_datetime = time.strftime("%Y%m%d_%H%M%S")

        # Utwórz nazwę pliku z datą i godziną, która pasuje do nazwy pliku w aplikacji Go
        filename = f"data/temperature_{current_datetime}.csv"
        filename = create_or_open_file(filename)

        # Zapisz temperaturę do pliku z właściwą datą i godziną
        with open(filename, "a") as log:
            log.write("{0},{1}\n".format(time.strftime("%Y-%m-%d %H:%M:%S"), str(temp)))

        time.sleep(1)  # Poczekaj 1 sekundę przed kolejnym odczytem
except KeyboardInterrupt:
    pass
