// Speech synthesis utility tailored for the Bashkir language

/**
 * Phonetically normalizes Bashkir unique letters for speech synthesis engines
 * to prevent TTS engines from skipping or misspelling them.
 */
export function normalizeBashkirForTTS(text: string): string {
  return text
    .replace(/[Әә]/g, 'э')
    .replace(/[Өө]/g, 'о')
    .replace(/[Үү]/g, 'у')
    .replace(/[Ғғ]/g, 'г')
    .replace(/[Ҡҡ]/g, 'к')
    .replace(/[Ңң]/g, 'нг')
    .replace(/[Ҙҙ]/g, 'з')
    .replace(/[Ҫҫ]/g, 'с')
    .replace(/[Һһ]/g, 'х');
}

/**
 * Plays a soft, harmonic bell chime using Web Audio API before speech
 */
export function playChime(): void {
  try {
    const AudioContextClass = window.AudioContext || (window as any).webkitAudioContext;
    if (!AudioContextClass) return;
    const ctx = new AudioContextClass();
    
    // Create warm kurai-like harmonic tone
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.type = 'sine';
    osc.frequency.setValueAtTime(587.33, ctx.currentTime); // D5 tone
    osc.frequency.exponentialRampToValueAtTime(880, ctx.currentTime + 0.12); // A5 chime

    gain.gain.setValueAtTime(0.08, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.25);

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.start();
    osc.stop(ctx.currentTime + 0.25);
  } catch (e) {
    // Ignore audio context autoplay limitations if blocked
  }
}

/**
 * Pronounces Bashkir text using Web Speech API with fallback and phonetic optimization
 */
export function speakBashkir(
  text: string,
  callbacks?: {
    onStart?: () => void;
    onEnd?: () => void;
    onError?: (err: any) => void;
  }
): void {
  if (!('speechSynthesis' in window)) {
    callbacks?.onError?.('Web Speech API не поддерживается вашим браузером');
    return;
  }

  // Cancel any ongoing speech
  window.speechSynthesis.cancel();

  // Play chime for responsiveness
  playChime();

  const phonetized = normalizeBashkirForTTS(text);
  const utterance = new SpeechSynthesisUtterance(phonetized);

  // Rate and pitch tuned for clear, melodic Bashkir pronunciation
  utterance.rate = 0.88;
  utterance.pitch = 1.05;

  // Search for best matching voice (Bashkir, Tatar, Turkish, Russian)
  const voices = window.speechSynthesis.getVoices();
  const preferredVoice =
    voices.find(v => v.lang.startsWith('ba') || v.lang.startsWith('tt')) ||
    voices.find(v => v.lang.startsWith('tr')) ||
    voices.find(v => v.lang.startsWith('ru') && !v.name.includes('Google')) ||
    voices.find(v => v.lang.startsWith('ru')) ||
    voices[0];

  if (preferredVoice) {
    utterance.voice = preferredVoice;
    utterance.lang = preferredVoice.lang;
  } else {
    utterance.lang = 'ru-RU';
  }

  utterance.onstart = () => {
    callbacks?.onStart?.();
  };

  utterance.onend = () => {
    callbacks?.onEnd?.();
  };

  utterance.onerror = (e) => {
    callbacks?.onEnd?.();
    callbacks?.onError?.(e);
  };

  // Small delay to let the chime sound first
  setTimeout(() => {
    window.speechSynthesis.speak(utterance);
  }, 120);
}

export function stopSpeaking(): void {
  if ('speechSynthesis' in window) {
    window.speechSynthesis.cancel();
  }
}
