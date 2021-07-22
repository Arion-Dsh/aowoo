
#include <memory>
#include <mutex>
#include <vector>

#include "aowoo_oboe.h"

#include "_cgo_export.h"
#include "oboe_oboe_Oboe.h"


namespace {


class Player: public oboe::AudioStreamDataCallback {
public:

	static Player &Get();

	int32_t Open(int rate, int bit_depth, int chs, int frames);
	int32_t Pause(int state);
	int32_t Close();
	oboe::DataCallbackResult onAudioReady(oboe::AudioStream
			*oboe_stream,
			void *audio_data,
			int32_t num_frames) override;

private:
	Player();
	int rate_ = 0;
	int chs_ = 0;
	int bit_depth = 0;
	int frames_= 0; // frames perdatacallback

    std::vector<float> buff_;
	std::mutex mLock;
	std::shared_ptr<oboe::AudioStream> mStream;
};

Player &Player::Get() {
  static Player *player = new Player();
  return *player;
}

Player::Player() = default;

int32_t Player::Open(int rate, int bit_depth, int chs, int frames) {
    rate_ = rate;
    chs_ = chs;
    bit_depth = bit_depth;
    frames_ = frames;

    std::vector<float> buf(frames_ *chs_);
    buff_ = buf;

    oboe::Result result;
    std::lock_guard<std::mutex> lock(mLock);
    oboe::AudioStreamBuilder builder;
    result = builder.setDirection(oboe::Direction::Output)
        ->setSharingMode(oboe::SharingMode::Exclusive)
        ->setPerformanceMode(oboe::PerformanceMode::LowLatency)
        ->setChannelCount(chs_)
        ->setSampleRate(rate_)
        ->setFramesPerDataCallback(frames_)
        ->setFormat(oboe::AudioFormat::Float)
        ->setDataCallback(this)
        ->openStream(mStream);

    if (result != oboe::Result::OK){
        return (int32_t) result;
    };
    result = mStream->start();
    if (result != oboe::Result::OK) {
        return (int32_t) result;
    }
    return 0;

}

int32_t Player::Pause(int state){
    std::lock_guard<std::mutex> lock(mLock);
    oboe::Result result;
    if (state == 1) {
        result = mStream ->pause();
    } else {
        result = mStream ->start();
    }
    return (int32_t) result;
}

int32_t Player::Close(){
	oboe::Result r;
	if (mStream) {
		r = mStream->stop();
		if (r != oboe::Result::OK) {
			return (int32_t) r;
		}
		r = mStream->close();
		if (r != oboe::Result::OK) {
			return (int32_t) r;
		}
		 mStream.reset();
	}
	return (int32_t) r;
}

oboe::DataCallbackResult Player::onAudioReady(oboe::AudioStream
        *oboe_stream,
        void *audio_data,
        int32_t num_frames){

    read_data(&buff_[0], num_frames * chs_);
    std::copy(buff_.begin(), buff_.end(),
            reinterpret_cast<float *>(audio_data));

    return oboe::DataCallbackResult::Continue;

 }


} //namespace

extern "C" {

static int aowoo_opened = 0;

int32_t aowoo_open_oboe(int rate, int bit_depth, int chs, int frames) {
     int32_t err = 0;
	if (aowoo_opened != 1) {
		err = Player::Get().Open(rate, bit_depth, chs, frames);
        if (err != 0) {
            return err;
        }
		aowoo_opened = 1;
	}
    return err;
}

int32_t aowoo_pause_oboe(int state) {
		return Player::Get().Pause(state);
}

}// extern "C"
