{ config, ... }:

{
  # Keymap in X11
  services.xserver.xkb = {
    layout = "us";
    variant = "";
  };

  # Audio
  security.rtkit.enable = true;
  services.pipewire = {
    enable = true;
    alsa.enable = true;
    alsa.support32Bit = true;
    pulse.enable = true;
  };

  # Bluetooth
  hardware.bluetooth.enable = true;
  hardware.bluetooth.powerOnBoot = true;
  services.blueman.enable = true;

  # Virtualisation
  virtualisation.podman.enable = true;
  virtualisation.oci-containers.backend = "podman";

  # Ollama
  services.ollama = {
    enable = true;
    environmentVariables = {
      OLLAMA_MODELS = config.pilo.ollama.modelsPath;
    };
  };
}