#include <windows.h>
#include <mmsystem.h>
#include <stdio.h>
#include <mmreg.h>
#include <Ksmedia.h>
#include <malloc.h>


#include "_cgo_export.h"

#define AOWOO_NUM_BUFFERS 3
#define AOWOO_BUFFER_SIZE 4096


HWAVEOUT hWaveOut;
WAVEHDR waveHeader[6];

void QueueData(HWAVEOUT waveOut, WAVEHDR *waveHeader);
void CALLBACK WaveOutProc(HWAVEOUT waveOut, UINT uMsg, DWORD_PTR dwInstance, DWORD_PTR dwParam1, DWORD_PTR dwParam2);

void initWaveEx(register WAVEFORMATEXTENSIBLE * wave, int rate, int depth, int chs)
{

	GUID subformatGuid = {STATIC_KSDATAFORMAT_SUBTYPE_IEEE_FLOAT};

	ZeroMemory(wave, sizeof(WAVEFORMATEXTENSIBLE));
	wave->Format.wFormatTag = WAVE_FORMAT_EXTENSIBLE;
	wave->Format.cbSize = 22;
	wave->Format.nChannels = chs;
	wave->Format.nSamplesPerSec = rate;
	wave->Format.wBitsPerSample = wave->Samples.wValidBitsPerSample = depth *chs;
	wave->Format.nBlockAlign = chs * (depth *chs /8);
	CopyMemory(&wave->SubFormat, &subformatGuid, sizeof(GUID));
	wave->Format.nAvgBytesPerSec = rate * wave->Format.nBlockAlign;
}



int aowoo_open(HWAVEOUT **waveOut, int rate, int depth, int chs, void *ctx)
{
    DWORD   err;

	WAVEFORMATEXTENSIBLE mixFormat;

	initWaveEx(&mixFormat, rate, depth, chs);

    err = waveOutOpen(&hWaveOut, WAVE_MAPPER, (WAVEFORMATEX*)&mixFormat, (DWORD_PTR)WaveOutProc, 0, CALLBACK_FUNCTION);
    *waveOut = &hWaveOut;

    if (err != MMSYSERR_NOERROR)
    {
        WCHAR errmsg[MAXERRORLENGTH + 1];
        MMRESULT converr;
        converr = waveOutGetErrorTextW(err, errmsg, MAXERRORLENGTH + 1);
        fwprintf(stderr, L"open error: %s\n", errmsg);
        return 1;
    }
	unsigned int i;
	for (i=0; i < 3; i++) {

		waveHeader[i].lpData = (char *)VirtualAlloc(0, AOWOO_BUFFER_SIZE, MEM_COMMIT, PAGE_READWRITE);
		/* waveHeader.lpData = (char *)calloc(AOWOO_BUFFER_SIZE, sizeof(float)); */
		waveHeader[i].dwBufferLength = AOWOO_BUFFER_SIZE;
		waveHeader[i].dwUser = (DWORD_PTR)ctx;
		err = waveOutPrepareHeader(hWaveOut, &waveHeader[i], sizeof(WAVEHDR));

		if (err != MMSYSERR_NOERROR)
		{
			WCHAR errmsg[MAXERRORLENGTH + 1];
			MMRESULT converr;
			converr = waveOutGetErrorTextW(err, errmsg, MAXERRORLENGTH + 1);
			fwprintf(stderr, L"open error: %s\n", errmsg);

		}

		QueueData(hWaveOut, &waveHeader[i]);

	}
}

void QueueData(HWAVEOUT waveOut, WAVEHDR *waveHeader){

		void *ctx_ = (void *)waveHeader->dwUser;
		float *casted_buffer = (float *)waveHeader->lpData;
        read_data(&casted_buffer[0], AOWOO_BUFFER_SIZE/4, ctx_);
		waveHeader->dwFlags &= ~WHDR_DONE;
        waveOutWrite(waveOut, waveHeader, sizeof(WAVEHDR));
}

void CALLBACK WaveOutProc(HWAVEOUT waveOut, UINT uMsg, DWORD_PTR dwInstance, DWORD_PTR dwParam1, DWORD_PTR dwParam2)
{

    if (uMsg == MM_WOM_DONE) {


        WAVEHDR *waveHeader = (WAVEHDR *)dwParam1;
       QueueData(waveOut, waveHeader);
    }
}


